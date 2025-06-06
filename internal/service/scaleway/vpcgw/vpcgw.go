package vpcgw

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/common"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// When capsManagedIPTag is set on a Public Gateway, its IP will be removed on
// Gateway deletion.
const capsManagedIPTag = "caps-vpcgw-ip=managed"

type Service struct {
	*scope.Cluster
}

func New(clusterScope *scope.Cluster) *Service {
	return &Service{Cluster: clusterScope}
}

func (s Service) Name() string {
	return "vpcgw"
}

func (s *Service) ensureGateways(ctx context.Context, delete bool) ([]*vpcgw.Gateway, error) {
	var desired []infrav1.PublicGatewaySpec
	// When delete is set, we ensure an empty list of Gateways to remove everything.
	if !delete {
		desired = s.ScalewayCluster.Spec.Network.PublicGateways
	}

	drle := &common.ResourceEnsurer[infrav1.PublicGatewaySpec, *vpcgw.Gateway]{
		ResourceReconciler: &desiredResourceListManager{s.Cluster},
	}
	return drle.Do(ctx, desired)
}

func (s *Service) ensureGatewaysAttachment(ctx context.Context, gateways []*vpcgw.Gateway, pnID string) error {
	for _, gateway := range gateways {
		if slices.ContainsFunc(gateway.GatewayNetworks, func(gn *vpcgw.GatewayNetwork) bool {
			return gn.PrivateNetworkID == pnID
		}) {
			continue
		}

		if gateway.Status != vpcgw.GatewayStatusRunning {
			return scaleway.WithTransientError(
				fmt.Errorf("gateway %s is not yet ready: currently %s", gateway.ID, gateway.Status),
				time.Second,
			)
		}

		if err := s.ScalewayClient.CreateGatewayNetwork(ctx, gateway.Zone, gateway.ID, pnID); err != nil {
			return fmt.Errorf("failed to create gateway network for gateway %s: %w", gateway.ID, err)
		}
	}

	return nil
}

func (s *Service) Reconcile(ctx context.Context) error {
	if !s.HasPrivateNetwork() {
		return nil
	}

	gateways, err := s.ensureGateways(ctx, false)
	if err != nil {
		return err
	}

	pnID, err := s.PrivateNetworkID()
	if err != nil {
		return err
	}

	if err := s.ensureGatewaysAttachment(ctx, gateways, pnID); err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context) error {
	if !s.HasPrivateNetwork() {
		return nil
	}

	_, err := s.ensureGateways(ctx, true)
	if err != nil {
		return err
	}

	return nil
}

type desiredResourceListManager struct {
	*scope.Cluster
}

func (d *desiredResourceListManager) ListResources(ctx context.Context) ([]*vpcgw.Gateway, error) {
	return d.ScalewayClient.FindGateways(ctx, d.ResourceTags())
}

func (d *desiredResourceListManager) DeleteResource(ctx context.Context, resource *vpcgw.Gateway) error {
	logf.FromContext(ctx).Info("Deleting Gateway", "gatewayName", resource.Name, "zone", resource.Zone)

	if err := d.ScalewayClient.DeleteGateway(
		ctx,
		resource.Zone,
		resource.ID,
		slices.Contains(resource.Tags, capsManagedIPTag),
	); err != nil {
		return fmt.Errorf("failed to delete Gateway: %w", err)
	}

	return nil
}

func (d *desiredResourceListManager) UpdateResource(
	ctx context.Context,
	resource *vpcgw.Gateway,
	desired infrav1.PublicGatewaySpec,
) (*vpcgw.Gateway, error) {
	return resource, nil
}

func (d *desiredResourceListManager) GetResourceZone(resource *vpcgw.Gateway) scw.Zone {
	return resource.Zone
}

func (d *desiredResourceListManager) GetResourceName(resource *vpcgw.Gateway) string {
	return resource.Name
}

func (d *desiredResourceListManager) GetDesiredZone(desired infrav1.PublicGatewaySpec) (scw.Zone, error) {
	return d.ScalewayClient.GetZoneOrDefault(desired.Zone)
}

func (d *desiredResourceListManager) ShouldKeepResource(
	resource *vpcgw.Gateway,
	desired infrav1.PublicGatewaySpec,
) bool {
	// Gateway has no IPv4, remove it and recreate it.
	if resource.IPv4 == nil {
		return false
	}

	if desired.Type != nil && *desired.Type != resource.Type {
		return false
	}

	if desired.IP == nil && !slices.Contains(resource.Tags, capsManagedIPTag) {
		return false
	}

	if desired.IP != nil && resource.IPv4.Address.String() != *desired.IP {
		return false
	}

	return true
}

func (d *desiredResourceListManager) GetDesiredResourceName(i int) string {
	return d.ResourceName(strconv.Itoa(i))
}

func (d *desiredResourceListManager) CreateResource(
	ctx context.Context,
	zone scw.Zone,
	name string,
	desired infrav1.PublicGatewaySpec,
) (*vpcgw.Gateway, error) {
	var ipID *string
	var gwType string
	tags := d.ResourceTags()

	if desired.IP != nil {
		ip, err := d.ScalewayClient.FindGatewayIP(ctx, zone, *desired.IP)
		if err != nil {
			if client.IsNotFoundError(err) {
				return nil, scaleway.WithTerminalError(fmt.Errorf("failed to find gateway ip: %w", err))
			}

			return nil, fmt.Errorf("failed to find gateway ip: %w", err)
		}

		ipID = &ip.ID
	} else {
		tags = append(tags, capsManagedIPTag)
	}

	if desired.Type != nil {
		gwType = *desired.Type
	}

	logf.FromContext(ctx).Info("Creating Gateway", "gatewayName", name, "zone", zone)

	gateway, err := d.ScalewayClient.CreateGateway(ctx, zone, name, gwType, tags, ipID)
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway: %w", err)
	}

	return gateway, nil
}
