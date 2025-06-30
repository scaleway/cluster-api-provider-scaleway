package lb

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/common"
	lbutil "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/lb/util"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// LB Tags.
	CAPSMainLBTag    = "caps-lb=main"
	CAPSExtraLBTag   = "caps-lb=extra"
	capsManagedIPTag = "caps-lb-ip=managed"

	// Backend port, must match port of apiservers.
	backendControlPlanePort = 6443

	BackendName  = "kube-apiserver"
	FrontendName = "kube-apiserver"

	// ACL indexes.
	aclIndex        = 0
	denyAllACLIndex = math.MaxInt32

	// ACL names.
	allowedRangesACLName = "allowed-ranges"
	publicGatewayACLName = "public-gateway"
	denyAllACLName       = "deny-all"
)

type Service struct {
	*scope.Cluster
}

func New(clusterScope *scope.Cluster) *Service {
	return &Service{Cluster: clusterScope}
}

func (s *Service) Name() string {
	return "lb"
}

func (s *Service) Reconcile(ctx context.Context) error {
	lb, err := s.ensureLB(ctx)
	if err != nil {
		return err
	}

	lbIP, err := getLBIPv4(lb)
	if err != nil {
		return err
	}

	s.SetStatusLoadBalancerIP(lbIP)

	extraLBs, err := s.ensureExtraLBs(ctx, false)
	if err != nil {
		return err
	}

	extraLBIPs := make([]string, 0, len(extraLBs))
	for _, extraLB := range extraLBs {
		extraLBIP, err := getLBIPv4(extraLB)
		if err != nil {
			return err
		}

		extraLBIPs = append(extraLBIPs, extraLBIP)
	}

	s.SetStatusExtraLoadBalancerIPs(extraLBIPs)

	if err := checkLBsReadiness(append(extraLBs, lb)); err != nil {
		return err
	}

	var pnID *string
	if s.HasPrivateNetwork() {
		tmpPNID, err := s.PrivateNetworkID()
		if err != nil {
			return err
		}

		pnID = &tmpPNID
	}

	if err := s.ensurePrivateNetwork(ctx, append(extraLBs, lb), pnID); err != nil {
		return err
	}

	backendByLB, err := s.ensureBackend(ctx, lb, extraLBs)
	if err != nil {
		return fmt.Errorf("failed to ensure lb backend: %w", err)
	}

	frontendByLB, err := s.ensureFrontend(ctx, backendByLB)
	if err != nil {
		return fmt.Errorf("failed to ensure lb frontend: %w", err)
	}

	if err := s.ensureACLs(ctx, lb, frontendByLB, pnID); err != nil {
		return fmt.Errorf("failed to ensure ACLs: %w", err)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context) error {
	if err := s.ensureDeleteLB(ctx); err != nil {
		return err
	}

	if _, err := s.ensureExtraLBs(ctx, true); err != nil {
		return err
	}

	return nil
}

func checkLBsReadiness(lbs []*lb.LB) error {
	for _, loadbalancer := range lbs {
		if loadbalancer.Status != lb.LBStatusReady {
			return scaleway.WithTransientError(
				fmt.Errorf("lb %s is not yet ready: currently %s", loadbalancer.ID, loadbalancer.Status),
				time.Second,
			)
		}
	}

	return nil
}

func (s *Service) ensureLB(ctx context.Context) (*lb.LB, error) {
	var spec infrav1.LoadBalancerSpec
	if s.ScalewayCluster.Spec.Network != nil && s.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer != nil {
		spec = s.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.LoadBalancerSpec
	}

	zone, lbType, err := lbutil.LBSpec(s.ScalewayClient, spec)
	if err != nil {
		return nil, err
	}

	lb, err := s.ScalewayClient.FindLB(ctx, zone, s.ResourceTags(CAPSMainLBTag))
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return nil, err
	}

	// If lb type does not match, migrate the LB.
	if lb != nil && !strings.EqualFold(lb.Type, lbType) {
		lb, err = s.ScalewayClient.MigrateLB(ctx, zone, lb.ID, lbType)
		if err != nil {
			return nil, fmt.Errorf("failed to migrate lb: %w", err)
		}
	} else if lb == nil {
		var ipID *string

		if spec.IP != nil {
			foundIP, err := s.ScalewayClient.FindLBIP(ctx, zone, *spec.IP)
			if err != nil {
				if client.IsNotFoundError(err) {
					return nil, scaleway.WithTerminalError(fmt.Errorf("failed to find IP %q: %w", *spec.IP, err))
				}

				return nil, fmt.Errorf("failed to find IP %q: %w", *spec.IP, err)
			}

			ipID = &foundIP.ID
		}

		logf.FromContext(ctx).Info("Creating main LB", "zone", zone)
		lb, err = s.ScalewayClient.CreateLB(ctx, zone, s.ResourceName(), lbType, ipID, s.ResourceTags(CAPSMainLBTag))
		if err != nil {
			return nil, err
		}
	}

	return lb, nil
}

func (s *Service) ensureDeleteLB(ctx context.Context) error {
	var spec infrav1.LoadBalancerSpec
	if s.ScalewayCluster.Spec.Network != nil && s.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer != nil {
		spec = s.ScalewayCluster.Spec.Network.ControlPlaneLoadBalancer.LoadBalancerSpec
	}

	zone, _, err := lbutil.LBSpec(s.ScalewayClient, spec)
	if err != nil {
		// If there is an error here, we can assume that no infra was created so there
		// is nothing to delete.
		return nil
	}

	lb, err := s.ScalewayClient.FindLB(ctx, zone, s.ResourceTags(CAPSMainLBTag))
	if err != nil {
		if errors.Is(err, client.ErrNoItemFound) {
			return nil
		}

		return err
	}

	logf.FromContext(ctx).Info("Deleting main LB")
	if err := s.ScalewayClient.DeleteLB(ctx, zone, lb.ID, spec.IP == nil); err != nil {
		return fmt.Errorf("failed to delete lb: %w", err)
	}

	return nil
}

func getLBIPv4(lb *lb.LB) (string, error) {
	for _, ip := range lb.IP {
		addr, err := netip.ParseAddr(ip.IPAddress)
		if err != nil {
			return "", err
		}

		if addr.Is4() {
			return ip.IPAddress, nil
		}
	}

	return "", fmt.Errorf("did not find ipv4 for lb %s", lb.ID)
}

func (s *Service) ensureExtraLBs(ctx context.Context, delete bool) ([]*lb.LB, error) {
	var desired []infrav1.LoadBalancerSpec
	// When delete is set, we ensure an empty list of LBs to remove everything.
	if !delete && s.ScalewayCluster.Spec.Network != nil {
		desired = s.ScalewayCluster.Spec.Network.ControlPlaneExtraLoadBalancers
	}

	drle := &common.ResourceEnsurer[infrav1.LoadBalancerSpec, *lb.LB]{
		ResourceReconciler: &desiredResourceListManager{s.Cluster},
	}
	return drle.Do(ctx, desired)
}

type desiredResourceListManager struct {
	*scope.Cluster
}

func (d *desiredResourceListManager) ListResources(ctx context.Context) ([]*lb.LB, error) {
	return d.ScalewayClient.FindLBs(ctx, d.ResourceTags(CAPSExtraLBTag))
}

func (d *desiredResourceListManager) DeleteResource(ctx context.Context, resource *lb.LB) error {
	logf.FromContext(ctx).Info("Deleting extra LB", "lbName", resource.Name, "zone", resource.Zone)

	if err := d.ScalewayClient.DeleteLB(
		ctx,
		resource.Zone,
		resource.ID,
		slices.Contains(resource.Tags, capsManagedIPTag),
	); err != nil {
		return fmt.Errorf("failed to delete LB: %w", err)
	}

	return nil
}

func (d *desiredResourceListManager) GetResourceZone(resource *lb.LB) scw.Zone {
	return resource.Zone
}

func (d *desiredResourceListManager) GetResourceName(resource *lb.LB) string {
	return resource.Name
}

func (d *desiredResourceListManager) GetDesiredZone(desired infrav1.LoadBalancerSpec) (scw.Zone, error) {
	return d.ScalewayClient.GetZoneOrDefault(desired.Zone)
}

func (d *desiredResourceListManager) ShouldKeepResource(
	_ context.Context,
	resource *lb.LB,
	desired infrav1.LoadBalancerSpec,
) (bool, error) {
	// If LB does not have an IP, remove it and recreate it.
	if len(resource.IP) == 0 {
		return false, nil
	}

	if desired.IP == nil && !slices.Contains(resource.Tags, capsManagedIPTag) {
		return false, nil
	}

	if desired.IP != nil && resource.IP[0].IPAddress != *desired.IP {
		return false, nil
	}

	return true, nil
}

func (d *desiredResourceListManager) GetDesiredResourceName(i int) string {
	return d.ResourceName(strconv.Itoa(i))
}

func (d *desiredResourceListManager) CreateResource(
	ctx context.Context,
	zone scw.Zone,
	name string,
	desired infrav1.LoadBalancerSpec,
) (*lb.LB, error) {
	_, lbType, err := lbutil.LBSpec(d.ScalewayClient, desired)
	if err != nil {
		return nil, err
	}

	tags := d.ResourceTags(CAPSExtraLBTag)

	var ipID *string
	if desired.IP != nil {
		foundIP, err := d.ScalewayClient.FindLBIP(ctx, zone, *desired.IP)
		if err != nil {
			if client.IsNotFoundError(err) {
				return nil, scaleway.WithTerminalError(fmt.Errorf("failed to find IP %q: %w", *desired.IP, err))
			}

			return nil, fmt.Errorf("failed to find IP %q: %w", *desired.IP, err)
		}

		ipID = &foundIP.ID
	} else {
		tags = append(tags, capsManagedIPTag)
	}

	logf.FromContext(ctx).Info("Creating extra LB", "lbName", name, "zone", zone)
	lb, err := d.ScalewayClient.CreateLB(ctx, zone, name, lbType, ipID, tags)
	if err != nil {
		return nil, err
	}
	return lb, nil
}

func (d *desiredResourceListManager) UpdateResource(
	ctx context.Context,
	resource *lb.LB,
	desired infrav1.LoadBalancerSpec,
) (*lb.LB, error) {
	if desired.Type != nil && !strings.EqualFold(*desired.Type, resource.Type) {
		logf.FromContext(ctx).Info("Migrating extra LB", "lbName", resource.Name, "zone", resource, "type", *desired.Type)
		return d.ScalewayClient.MigrateLB(ctx, resource.Zone, resource.ID, *desired.Type)
	}

	return resource, nil
}

func (s *Service) getOrCreateBackend(
	ctx context.Context,
	lb *lb.LB,
	servers []string,
	updateServers bool,
) (*lb.Backend, error) {
	servers = slices.Sorted(slices.Values(servers))

	backend, err := s.ScalewayClient.FindBackend(ctx, lb.Zone, lb.ID, BackendName)
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return nil, err
	}

	if backend == nil {
		backend, err = s.ScalewayClient.CreateBackend(
			ctx,
			lb.Zone,
			lb.ID,
			BackendName,
			servers,
			backendControlPlanePort,
		)
		if err != nil {
			return nil, err
		}
	} else if updateServers && !slices.Equal(servers, slices.Sorted(slices.Values(backend.Pool))) {
		backend, err = s.ScalewayClient.SetBackendServers(ctx, lb.Zone, backend.ID, servers)
		if err != nil {
			return nil, err
		}
	}

	return backend, nil
}

func (s *Service) ensureBackend(ctx context.Context, mainLB *lb.LB, extraLBs []*lb.LB) ([]*lb.Backend, error) {
	backends := make([]*lb.Backend, 0, len(extraLBs)+1)

	mainLBBackend, err := s.getOrCreateBackend(ctx, mainLB, nil, false)
	if err != nil {
		return nil, err
	}

	backends = append(backends, mainLBBackend)

	for _, extraLB := range extraLBs {
		backend, err := s.getOrCreateBackend(ctx, extraLB, mainLBBackend.Pool, true)
		if err != nil {
			return nil, err
		}

		backends = append(backends, backend)
	}

	return backends, nil
}

func (s *Service) ensureFrontend(ctx context.Context, backends []*lb.Backend) (map[string]*lb.Frontend, error) {
	frontendByLB := make(map[string]*lb.Frontend)

	for _, backend := range backends {
		frontend, err := s.ScalewayClient.FindFrontend(
			ctx, backend.LB.Zone,
			backend.LB.ID,
			FrontendName,
		)
		if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
			return nil, err
		}

		if frontend == nil {
			frontend, err = s.ScalewayClient.CreateFrontend(
				ctx,
				backend.LB.Zone,
				backend.LB.ID,
				FrontendName,
				backend.ID,
				s.ControlPlaneLoadBalancerPort(),
			)
			if err != nil {
				return nil, err
			}
		}

		frontendByLB[backend.LB.ID] = frontend
	}

	return frontendByLB, nil
}

func (s *Service) ensurePrivateNetwork(ctx context.Context, lbs []*lb.LB, pnID *string) error {
	if pnID == nil {
		return nil
	}

	for _, lb := range lbs {
		lbPN, err := s.ScalewayClient.FindLBPrivateNetwork(ctx, lb.Zone, lb.ID, *pnID)
		if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
			return err
		}

		if lbPN == nil {
			if err := s.ScalewayClient.AttachLBPrivateNetwork(ctx, lb.Zone, lb.ID, *pnID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) ensureACLs(
	ctx context.Context,
	mainLB *lb.LB,
	frontendByLB map[string]*lb.Frontend,
	pnID *string,
) error {
	allowedRanges := s.ControlPlaneLoadBalancerAllowedRanges()

	var denyAll []string
	if len(allowedRanges) > 0 {
		denyAll = []string{"0.0.0.0/0", "::/0"}
	}

	var publicGatewayIPs []string
	if pnID != nil && s.HasPrivateNetwork() {
		gws, err := s.ScalewayClient.FindGateways(ctx, s.ResourceTags())
		if err != nil {
			return err
		}

		for _, gw := range gws {
			if gw.IPv4 != nil {
				publicGatewayIPs = append(publicGatewayIPs, gw.IPv4.Address.String())
			}
		}
	}

	mainLBFrontend := frontendByLB[mainLB.ID]
	if mainLBFrontend == nil {
		panic("did not expect mainLBFrontend to be nil")
	}

	// Set the Allowed Ranges ACL.
	if err := s.ensureACL(ctx, mainLBFrontend, allowedRangesACLName, allowedRanges, false, aclIndex); err != nil {
		return fmt.Errorf("failed to ensure %s ACL: %w", allowedRangesACLName, err)
	}

	// Set the Public Gateway ACL.
	if err := s.ensureACL(ctx, mainLBFrontend, publicGatewayACLName, publicGatewayIPs, false, aclIndex); err != nil {
		return fmt.Errorf("failed to ensure %s ACL: %w", publicGatewayACLName, err)
	}

	// Set the Deny All ACL. If denyAll is empty, it will not be created (or it
	// will be deleted if it exists).
	if err := s.ensureACL(ctx, mainLBFrontend, denyAllACLName, denyAll, true, denyAllACLIndex); err != nil {
		return fmt.Errorf("failed to ensure %s ACL: %w", denyAllACLName, err)
	}

	if len(frontendByLB) > 1 {
		mainLBACLs, err := s.ScalewayClient.ListLBACLs(ctx, mainLB.Zone, mainLBFrontend.ID)
		if err != nil {
			return fmt.Errorf("failed to list ACLs: %w", err)
		}

		for id, frontend := range frontendByLB {
			if id == mainLB.ID {
				continue
			}

			extraLBACLs, err := s.ScalewayClient.ListLBACLs(ctx, frontend.LB.Zone, frontend.ID)
			if err != nil {
				return fmt.Errorf("failed to list ACLs for extra LB: %w", err)
			}

			if lbutil.ACLEqual(mainLBACLs, extraLBACLs) {
				continue
			}

			// Mismatch, let's correct it.
			if err := s.ScalewayClient.SetLBACLs(ctx, frontend.LB.Zone, frontend.ID, aclsToACLSpecs(mainLBACLs)); err != nil {
				return fmt.Errorf("failed to set acls: %w", err)
			}
		}
	}

	return nil
}

// ensureACL ensures the ACL with specified parameters exists or doesn't exist if
// the ACL doesn't contain any IP.
func (s *Service) ensureACL(
	ctx context.Context,
	frontend *lb.Frontend,
	name string,
	ips []string,
	deny bool,
	index int32,
) error {
	acl, err := s.ScalewayClient.FindLBACLByName(ctx, frontend.LB.Zone, frontend.ID, name)
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return err
	}

	// Remove ACL / Do nothing if there is no IP in it.
	if len(ips) == 0 {
		if acl != nil {
			if err := s.ScalewayClient.DeleteLBACL(ctx, frontend.LB.Zone, acl.ID); err != nil {
				return err
			}
		}

		return nil
	}

	action := lb.ACLActionTypeAllow
	if deny {
		action = lb.ACLActionTypeDeny
	}

	// Create ACL if it does not exist.
	if acl == nil {
		return s.ScalewayClient.CreateLBACL(ctx, frontend.LB.Zone, frontend.ID, name, index, action, ips)
	}

	// Update ACL if ips are different.
	if acl.Match == nil || !lbutil.IPsEqual(scw.StringSlicePtr(ips), acl.Match.IPSubnet) {
		return s.ScalewayClient.UpdateLBACL(ctx, frontend.LB.Zone, acl.ID, name, index, action, ips)
	}

	return nil
}

func aclsToACLSpecs(acls []*lb.ACL) []*lb.ACLSpec {
	specs := make([]*lb.ACLSpec, 0, len(acls))

	for _, acl := range acls {
		specs = append(specs, &lb.ACLSpec{
			Name:        acl.Name,
			Action:      acl.Action,
			Match:       acl.Match,
			Index:       acl.Index,
			Description: acl.Description,
		})
	}

	return specs
}
