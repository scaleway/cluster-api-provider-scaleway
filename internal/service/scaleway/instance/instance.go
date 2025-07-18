package instance

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"text/template"
	"time"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	servicelb "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/lb"
	lbutil "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/lb/util"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// defaultRootVolumeSize is the default IOPS of the root block volume when it's
	// created with the instance API. If the user sets a different value, the volume
	// will be updated to the desired value.
	defaultRootVolumeIOPS = 5000
	// machineACLIndex is the index of the ACL to use for the machine.
	machineACLIndex = int32(1)
	// cloudInitUserDataKey is the key used to store the cloud-init user data in the server.
	cloudInitUserDataKey = "cloud-init"
)

// instanceVolumeTypeToMarketplaceType maps the instance volume type to the marketplace image type.
var instanceVolumeTypeToMarketplaceType = map[instance.VolumeVolumeType]marketplace.LocalImageType{
	instance.VolumeVolumeTypeLSSD:      marketplace.LocalImageTypeInstanceLocal,
	instance.VolumeVolumeTypeSbsVolume: marketplace.LocalImageTypeInstanceSbs,
}

type Service struct {
	*scope.Machine
}

func New(machineScope *scope.Machine) *Service {
	return &Service{Machine: machineScope}
}

func (s *Service) Name() string {
	return "instance"
}

func (s *Service) Reconcile(ctx context.Context) error {
	server, err := s.ensureServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure server: %w", err)
	}

	// Ensure the server configuration when the node has never joined the cluster.
	if !s.HasJoinedCluster() {
		server, err = s.ensurePublicIPs(ctx, server)
		if err != nil {
			return err
		}

		privateIPs, err := s.ensurePrivateNIC(ctx, server)
		if err != nil {
			return fmt.Errorf("failed to ensure private nic: %w", err)
		}

		lbs, err := s.findControlPlaneLBs(ctx)
		if err != nil {
			return err
		}

		nodeIP, err := nodeIP(server, privateIPs)
		if err != nil {
			return err
		}

		if err := s.ensureControlPlaneLBs(ctx, lbs, nodeIP, false); err != nil {
			return fmt.Errorf("failed to ensure control-plane lbs: %w", err)
		}

		if err := s.ensureControlPlaneLBsACL(ctx, lbs, instanceIPsToStrings(server.PublicIPs), false); err != nil {
			return fmt.Errorf("failed to ensure control-plane lbs acls: %w", err)
		}

		if err := s.ensureCloudInit(ctx, server, nodeIP); err != nil {
			return fmt.Errorf("failed to ensure cloud-init: %w", err)
		}

		s.SetProviderID(providerID(server))
		s.SetAddresses(machineAddresses(server, privateIPs))

		if err := s.ensureServerStarted(ctx, server); err != nil {
			return fmt.Errorf("failed to ensure server started: %w", err)
		}

		return nil
	}

	// The node has already joined the cluster, we can safely remove cloud init userdata.
	if err := s.ensureNoCloudInit(ctx, server); err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context) error {
	zone, err := s.Zone()
	if err != nil {
		// If zone is invalid, it's highly probable that nothing was provisioned.
		return nil
	}

	server, err := s.ScalewayClient.FindServer(ctx, zone, s.ResourceTags())
	if err != nil {
		if client.IsNotFoundError(err) {
			return nil
		}

		return err
	}

	lbs, err := s.findControlPlaneLBs(ctx)
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return err
	}

	if err := s.ensureControlPlaneLBsACL(ctx, lbs, nil, true); err != nil && !client.IsNotFoundError(err) {
		return fmt.Errorf("failed to ensure control-plane lbs acls: %w", err)
	}

	// Remove this control-plane from the loadbalancer.
	if s.IsControlPlane() {
		privateIPs, err := s.ensurePrivateNIC(ctx, server)
		if err != nil {
			return fmt.Errorf("failed to ensure private nic: %w", err)
		}

		// nodeIP's error is ignored as it means the server no longer has an IP.
		if nodeIP, err := nodeIP(server, privateIPs); err == nil {
			if err := s.ensureControlPlaneLBs(ctx, lbs, nodeIP, true); err != nil {
				return fmt.Errorf("failed to ensure control-plane lbs: %w", err)
			}
		}
	}

	if err := s.ensureNoPublicIPs(ctx, server); err != nil {
		return err
	}

	if err := s.ensureServerStopped(ctx, server); err != nil {
		return err
	}

	if err := s.ensureBootVolumeDeleted(ctx, server); err != nil {
		return err
	}

	if err := s.ScalewayClient.DeleteServer(ctx, zone, server.ID); err != nil {
		return err
	}

	return nil
}

func (s *Service) ensureServer(ctx context.Context) (*instance.Server, error) {
	zone, err := s.Zone()
	if err != nil {
		return nil, err
	}

	if server, err := s.ScalewayClient.FindServer(ctx, zone, s.ResourceTags()); err == nil {
		return server, nil
	} else if !client.IsNotFoundError(err) {
		return nil, err
	}

	// Provider ID is already set, it's not normal that we didn't find the server.
	if s.ScalewayMachine.Spec.ProviderID != nil {
		return nil, scaleway.WithTerminalError(fmt.Errorf("providerID is already set on ScalewayMachine, but no existing server was found"))
	}

	// Server does not exist, let's create it.
	logf.FromContext(ctx).Info("Creating instance server", "serverName", s.ResourceName(), "zone", zone)

	// First, find an image ID.
	volumeType, err := s.RootVolumeType()
	if err != nil {
		return nil, err
	}

	var imageID string
	switch image := s.ScalewayMachine.Spec.Image; {
	case image.ID != nil:
		imageID = *image.ID
	case image.Label != nil:
		marketplaceType, ok := instanceVolumeTypeToMarketplaceType[volumeType]
		if !ok {
			return nil, scaleway.WithTerminalError(fmt.Errorf("did not find marketplace type for volume type %s", volumeType))
		}

		image, err := s.ScalewayClient.GetLocalImageByLabel(ctx, zone, s.ScalewayMachine.Spec.CommercialType, *image.Label, marketplaceType)
		if err != nil {
			return nil, err
		}

		imageID = image.ID
	case image.Name != nil:
		image, err := s.ScalewayClient.FindImage(ctx, zone, *image.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to find image by name, make sure it exists in zone %s: %w", zone, err)
		}

		imageID = image.ID
	}

	if imageID == "" {
		return nil, errors.New("unable to find a valid image in ScalewayMachine spec")
	}

	placementGroupID, err := s.placementGroupID(ctx, zone)
	if err != nil {
		return nil, err
	}

	securityGroupID, err := s.securityGroupID(ctx, zone)
	if err != nil {
		return nil, err
	}

	// Finally, create the server.
	server, err := s.ScalewayClient.CreateServer(
		ctx,
		zone,
		s.ResourceName(),
		s.ScalewayMachine.Spec.CommercialType,
		imageID,
		placementGroupID,
		securityGroupID,
		s.RootVolumeSize(),
		volumeType,
		s.ResourceTags(),
	)
	if err != nil {
		return nil, err
	}

	// If server is created with an SBS root volume, we check if it's needed to update its IOPS.
	if rootVolume, ok := server.Volumes["0"]; ok && rootVolume.VolumeType == instance.VolumeServerVolumeTypeSbsVolume {
		desiredIOPS := s.RootVolumeIOPS()

		// We assume volumes created by instance have 5000 IOPS (defaultRootVolumeIOPS) by default.
		if desiredIOPS != nil && *desiredIOPS != defaultRootVolumeIOPS {
			if err := s.ScalewayClient.UpdateVolumeIOPS(ctx, zone, rootVolume.ID, *desiredIOPS); err != nil {
				return nil, fmt.Errorf("failed to update root volume iops: %w", err)
			}
		}
	}

	return server, nil
}

func (s *Service) placementGroupID(ctx context.Context, zone scw.Zone) (*string, error) {
	// If user has specified a placement group, get its ID.
	if s.ScalewayMachine.Spec.PlacementGroup != nil {
		switch pgref := s.ScalewayMachine.Spec.PlacementGroup; {
		case pgref.ID != nil:
			return pgref.ID, nil
		case pgref.Name != nil:
			placementGroup, err := s.ScalewayClient.FindPlacementGroup(ctx, zone, *pgref.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to find placement group: %w", err)
			}

			return &placementGroup.ID, nil
		}
	}

	return nil, nil
}

func (s *Service) securityGroupID(ctx context.Context, zone scw.Zone) (*string, error) {
	// If user has specified a security group, get its ID.
	if s.ScalewayMachine.Spec.SecurityGroup != nil {
		switch sgref := s.ScalewayMachine.Spec.SecurityGroup; {
		case sgref.ID != nil:
			return sgref.ID, nil
		case sgref.Name != nil:
			securityGroup, err := s.ScalewayClient.FindSecurityGroup(ctx, zone, *sgref.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to find security group: %w", err)
			}

			return &securityGroup.ID, nil
		}
	}

	return nil, nil
}

func (s *Service) ensurePublicIPs(ctx context.Context, server *instance.Server) (*instance.Server, error) {
	if !s.HasPublicIPv4() && !s.HasPublicIPv6() {
		return server, nil
	}

	ips, err := s.ScalewayClient.FindIPs(ctx, server.Zone, s.ResourceTags())
	if err != nil {
		return nil, err
	}

	publicIPIDs := []string{}
	updateServer := false

	for _, version := range []struct {
		ipType instance.IPType
		want   bool
	}{
		{ipType: instance.IPTypeRoutedIPv4, want: s.HasPublicIPv4()},
		{ipType: instance.IPTypeRoutedIPv6, want: s.HasPublicIPv6()},
	} {
		// Skip if we don't want this type of IP.
		if !version.want {
			continue
		}

		// Skip if IP already exists.
		ipIndex := slices.IndexFunc(ips, func(ip *instance.IP) bool { return ip.Type == version.ipType })
		if ipIndex != -1 {
			if ips[ipIndex].Server == nil {
				updateServer = true
			} else if ips[ipIndex].Server.ID != server.ID {
				return nil, fmt.Errorf("expected IP %s to be attached to %s", ips[ipIndex].ID, server.ID)
			}

			publicIPIDs = append(publicIPIDs, ips[ipIndex].ID)
			continue
		}

		ip, err := s.ScalewayClient.CreateIP(ctx, server.Zone, version.ipType, s.ResourceTags())
		if err != nil {
			return nil, fmt.Errorf("failed to create IP: %w", err)
		}

		publicIPIDs = append(publicIPIDs, ip.ID)
		updateServer = true
	}

	if updateServer {
		server, err = s.ScalewayClient.UpdateServerPublicIPs(ctx, server.Zone, server.ID, publicIPIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh server after updating IPs")
		}
	}

	return server, nil
}

func (s *Service) ensurePrivateNIC(ctx context.Context, server *instance.Server) ([]*ipam.IP, error) {
	if !s.HasPrivateNetwork() {
		return nil, nil
	}

	privateNetworkID, err := s.PrivateNetworkID()
	if err != nil {
		return nil, err
	}

	// Do nothing if Private NIC is already present.
	pnicIndex := slices.IndexFunc(server.PrivateNics, func(pnic *instance.PrivateNIC) bool {
		return pnic.PrivateNetworkID == privateNetworkID
	})

	var pnic *instance.PrivateNIC
	if pnicIndex == -1 {
		pnic, err = s.ScalewayClient.CreatePrivateNIC(ctx, server.Zone, server.ID, privateNetworkID)
		if err != nil {
			return nil, err
		}
	} else {
		pnic = server.PrivateNics[pnicIndex]
	}

	privateIPs, err := s.ScalewayClient.FindPrivateNICIPs(ctx, pnic.ID)
	if err != nil {
		return nil, err
	}

	if len(privateIPs) == 0 {
		return nil, scaleway.WithTransientError(errors.New("no private IP available in IPAM yet"), time.Second)
	}

	return privateIPs, nil
}

func machineAddresses(server *instance.Server, privateIPs []*ipam.IP) []clusterv1.MachineAddress {
	// The total number of addresses is len(server.PublicIPs) + len(privateIPs) + ExternalDNS + Hostname.
	addresses := make([]clusterv1.MachineAddress, 0, len(server.PublicIPs)+len(privateIPs)+2)

	addresses = append(addresses, clusterv1.MachineAddress{
		Type:    clusterv1.MachineHostName,
		Address: server.Hostname,
	})

	for _, publicIP := range server.PublicIPs {
		addresses = append(addresses, clusterv1.MachineAddress{
			Type:    clusterv1.MachineExternalIP,
			Address: publicIP.Address.String(),
		})
	}

	if len(server.PublicIPs) > 0 {
		addresses = append(addresses, clusterv1.MachineAddress{
			Type:    clusterv1.MachineExternalDNS,
			Address: fmt.Sprintf("%s.pub.instances.scw.cloud", server.ID),
		})
	}

	for _, privateIP := range privateIPs {
		addresses = append(addresses, clusterv1.MachineAddress{
			Type:    clusterv1.MachineInternalIP,
			Address: privateIP.Address.IP.String(),
		})
	}

	return addresses
}

func providerID(server *instance.Server) string {
	return fmt.Sprintf("scaleway://instance/%s/%s", server.Zone, server.ID)
}

func nodeIP(server *instance.Server, privateIPs []*ipam.IP) (string, error) {
	if len(privateIPs) > 0 {
		v4Index := slices.IndexFunc(privateIPs, func(ip *ipam.IP) bool { return !ip.IsIPv6 })
		if v4Index == -1 {
			return "", errors.New("did not find a Private IPv4")
		}

		return privateIPs[v4Index].Address.IP.String(), nil
	}

	v4Index := slices.IndexFunc(server.PublicIPs, func(ip *instance.ServerIP) bool { return ip.Family == instance.ServerIPIPFamilyInet })
	if v4Index == -1 {
		return "", errors.New("did not find a Public IPv4")
	}

	return server.PublicIPs[v4Index].Address.String(), nil
}

func (s *Service) findControlPlaneLBs(ctx context.Context) ([]*lb.LB, error) {
	var spec infrav1.LoadBalancerSpec
	if s.ScalewayCluster.Spec.Network != nil && s.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer != nil {
		spec = s.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.LoadBalancerSpec
	}

	zone, err := s.ScalewayClient.GetZoneOrDefault(spec.Zone)
	if err != nil {
		return nil, err
	}

	mainLB, err := s.ScalewayClient.FindLB(ctx, zone, s.Cluster.ResourceTags(servicelb.CAPSMainLBTag))
	if err != nil {
		return nil, err
	}

	extraLBs, err := s.ScalewayClient.FindLBs(ctx, s.Cluster.ResourceTags(servicelb.CAPSExtraLBTag))
	if err != nil {
		return nil, err
	}

	return append(extraLBs, mainLB), nil
}

func (s *Service) ensureControlPlaneLBs(ctx context.Context, lbs []*lb.LB, nodeIP string, deletion bool) error {
	if !s.IsControlPlane() {
		return nil
	}

	for _, loadbalancer := range lbs {
		if loadbalancer.Status == lb.LBStatusDeleting {
			continue
		}

		backend, err := s.ScalewayClient.FindBackend(ctx, loadbalancer.Zone, loadbalancer.ID, servicelb.BackendName)
		if err != nil {
			return err
		}

		switch {
		case deletion && slices.Contains(backend.Pool, nodeIP):
			if err := s.ScalewayClient.RemoveBackendServer(ctx, loadbalancer.Zone, backend.ID, nodeIP); err != nil {
				return err
			}
		case !deletion && !slices.Contains(backend.Pool, nodeIP):
			if err := s.ScalewayClient.AddBackendServer(ctx, loadbalancer.Zone, backend.ID, nodeIP); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) ensureControlPlaneLBsACL(ctx context.Context, lbs []*lb.LB, publicIPs []string, delete bool) error {
	for _, loadbalancer := range lbs {
		if loadbalancer.Status == lb.LBStatusDeleting {
			continue
		}

		frontend, err := s.ScalewayClient.FindFrontend(ctx, loadbalancer.Zone, loadbalancer.ID, servicelb.FrontendName)
		if err != nil {
			// If the frontend is not found, we can skip it when reconciling a deletion.
			if delete && client.IsNotFoundError(err) {
				continue
			}

			return err
		}

		acl, err := s.ScalewayClient.FindLBACLByName(ctx, loadbalancer.Zone, frontend.ID, s.ResourceName())
		if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
			return err
		}

		// If no publicIP is set, we either delete the existing ACL or do nothing.
		if len(publicIPs) == 0 {
			if acl != nil {
				if err := s.ScalewayClient.DeleteLBACL(ctx, loadbalancer.Zone, acl.ID); err != nil {
					return err
				}
			}

			continue
		}

		if acl == nil {
			if err := s.ScalewayClient.CreateLBACL(ctx,
				loadbalancer.Zone,
				frontend.ID,
				s.ResourceName(),
				machineACLIndex,
				lb.ACLActionTypeAllow,
				publicIPs,
			); err != nil {
				return err
			}

			continue
		}

		if acl.Match == nil || !lbutil.IPsEqual(acl.Match.IPSubnet, scw.StringSlicePtr(publicIPs)) {
			if err := s.ScalewayClient.UpdateLBACL(
				ctx,
				loadbalancer.Zone,
				acl.ID,
				s.ResourceName(),
				machineACLIndex,
				lb.ACLActionTypeAllow,
				publicIPs,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func instanceIPsToStrings(ips []*instance.ServerIP) []string {
	out := make([]string, 0, len(ips))

	for _, ip := range ips {
		out = append(out, ip.Address.String())
	}

	return out
}

func (s *Service) ensureCloudInit(ctx context.Context, server *instance.Server, nodeIP string) error {
	if server.State != instance.ServerStateStopped {
		return nil
	}

	userData, err := s.ScalewayClient.GetAllServerUserData(ctx, server.Zone, server.ID)
	if err != nil {
		return err
	}

	if _, ok := userData[cloudInitUserDataKey]; !ok {
		bootstrapData, err := s.GetBootstrapData(ctx)
		if err != nil {
			return err
		}

		// Apply custom templating on cloud-init bootstrap data.
		tmpl, err := template.New("").Delims("[[[", "]]]").Parse(string(bootstrapData))
		if err != nil {
			return fmt.Errorf("failed to parse bootstrap data as template: %w", err)
		}

		tmplExec := &strings.Builder{} // tmplExec will contain the executed template.
		tmplData := struct{ NodeIP string }{nodeIP}

		if err := tmpl.ExecuteTemplate(tmplExec, "", tmplData); err != nil {
			return fmt.Errorf("failed to execute bootstrap data template: %w", err)
		}

		if err := s.ScalewayClient.SetServerUserData(
			ctx,
			server.Zone,
			server.ID,
			cloudInitUserDataKey,
			tmplExec.String(),
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ensureNoCloudInit(ctx context.Context, server *instance.Server) error {
	userData, err := s.ScalewayClient.GetAllServerUserData(ctx, server.Zone, server.ID)
	if err != nil {
		return err
	}

	if _, ok := userData["cloud-init"]; !ok {
		return nil
	}

	return s.ScalewayClient.DeleteServerUserData(ctx, server.Zone, server.ID, cloudInitUserDataKey)
}

func (s *Service) ensureServerStarted(ctx context.Context, server *instance.Server) error {
	if server.State != instance.ServerStateStopped {
		return nil
	}

	return s.ScalewayClient.ServerAction(ctx, server.Zone, server.ID, instance.ServerActionPoweron)
}

func (s *Service) ensureServerStopped(ctx context.Context, server *instance.Server) error {
	if server.State == instance.ServerStateStopped {
		return nil
	}

	if server.State != instance.ServerStateStopping {
		if err := s.ScalewayClient.ServerAction(ctx, server.Zone, server.ID, instance.ServerActionPoweroff); err != nil {
			return err
		}
	}

	return scaleway.WithTransientError(errors.New("server is not stopped yet"), 10*time.Second)
}

func (s *Service) ensureNoPublicIPs(ctx context.Context, server *instance.Server) error {
	ips, err := s.ScalewayClient.FindIPs(ctx, server.Zone, s.ResourceTags())
	if err != nil {
		return err
	}

	for _, ip := range ips {
		if err := s.ScalewayClient.DeleteIP(ctx, server.Zone, ip.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ensureBootVolumeDeleted(ctx context.Context, server *instance.Server) error {
	// First: detach the boot volume
	for _, vol := range server.Volumes {
		if !vol.Boot {
			continue
		}

		// We add a tag on the boot volume to be able to find it and delete it in the next step.
		switch vol.VolumeType {
		case instance.VolumeServerVolumeTypeSbsVolume:
			if err := s.ScalewayClient.UpdateVolumeTags(ctx, server.Zone, vol.ID, s.ResourceTags()); err != nil {
				return err
			}
		case instance.VolumeServerVolumeTypeLSSD:
			if err := s.ScalewayClient.UpdateInstanceVolumeTags(ctx, server.Zone, vol.ID, s.ResourceTags()); err != nil {
				return err
			}
		default:
			return scaleway.WithTerminalError(fmt.Errorf("cannot detach unsupported boot volume with type %s", vol.VolumeType))
		}

		if err := s.ScalewayClient.DetachVolume(ctx, server.Zone, vol.ID); err != nil {
			return err
		}
	}

	// Finally: delete the volume. From here, we may no longer have the information
	// about the root volume type (l_ssd or sbs_volume), so we have to try both APIs.
	volume, err := s.ScalewayClient.FindVolume(ctx, server.Zone, s.ResourceTags())
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return err
	}

	if volume != nil {
		if volume.Status != block.VolumeStatusAvailable {
			return scaleway.WithTransientError(fmt.Errorf("root block volume is not yet ready to be deleted (%s)", volume.Status), 2*time.Second)
		}

		return s.ScalewayClient.DeleteVolume(ctx, server.Zone, volume.ID)
	}

	instanceVolume, err := s.ScalewayClient.FindInstanceVolume(ctx, server.Zone, s.ResourceTags())
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return err
	}

	if instanceVolume != nil {
		if instanceVolume.State != instance.VolumeStateAvailable {
			return scaleway.WithTransientError(fmt.Errorf("root volume is not yet ready to be deleted (%s)", instanceVolume.State), time.Second)
		}

		return s.ScalewayClient.DeleteInstanceVolume(ctx, server.Zone, instanceVolume.ID)
	}

	return nil
}
