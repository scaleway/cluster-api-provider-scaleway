package vpcgw

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/util/conditions"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/common"
)

// When capsManagedIPTag is set on a Public Gateway, its IP will be removed on
// Gateway deletion.
const capsManagedIPTag = "caps-vpcgw-ip=managed"

type Scope interface {
	scope.Interface
	conditions.Setter

	HasPrivateNetwork() bool
	PrivateNetworkID() (string, error)
	PublicGateways() []infrav1.PublicGateway
}
type Service struct {
	Scope
}

func New(s Scope) *Service {
	return &Service{s}
}

func (s Service) Name() string {
	return "vpcgw"
}

func (s *Service) ensureGateways(ctx context.Context, delete bool) ([]*vpcgw.Gateway, error) {
	var desired []infrav1.PublicGateway
	// When delete is set, we ensure an empty list of Gateways to remove everything.
	if !delete {
		desired = s.PublicGateways()
	}

	drle := &common.ResourceEnsurer[infrav1.PublicGateway, *vpcgw.Gateway]{
		ResourceReconciler: &desiredResourceListManager{s.Scope, make(map[scw.Zone][]string)},
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

		if err := s.Cloud().CreateGatewayNetwork(ctx, gateway.Zone, gateway.ID, pnID); err != nil {
			return fmt.Errorf("failed to create gateway network for gateway %s: %w", gateway.ID, err)
		}
	}

	return nil
}

func (s *Service) Reconcile(ctx context.Context) error {
	if !s.HasPrivateNetwork() {
		conditions.Set(s, metav1.Condition{
			Type:   infrav1.PublicGatewaysReadyCondition,
			Status: metav1.ConditionTrue,
			Reason: infrav1.NoPrivateNetworkReason,
		})
		return nil
	}

	gateways, err := s.ensureGateways(ctx, false)
	if err != nil {
		conditions.Set(s, metav1.Condition{
			Type:    infrav1.PublicGatewaysReadyCondition,
			Status:  metav1.ConditionFalse,
			Reason:  infrav1.ReconciliationFailedReason,
			Message: err.Error(),
		})
		return err
	}

	pnID, err := s.PrivateNetworkID()
	if err != nil {
		return err
	}

	if err := s.ensureGatewaysAttachment(ctx, gateways, pnID); err != nil {
		conditions.Set(s, metav1.Condition{
			Type:    infrav1.PublicGatewaysReadyCondition,
			Status:  metav1.ConditionFalse,
			Reason:  infrav1.PrivateNetworkAttachmentFailedReason,
			Message: err.Error(),
		})
		return err
	}

	conditions.Set(s, metav1.Condition{
		Type:   infrav1.PublicGatewaysReadyCondition,
		Status: metav1.ConditionTrue,
		Reason: infrav1.ProvisionedReason,
	})

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
	Scope

	gatewayTypesCache map[scw.Zone][]string
}

func (d *desiredResourceListManager) ListResources(ctx context.Context) ([]*vpcgw.Gateway, error) {
	return d.Cloud().FindGateways(ctx, d.ResourceTags())
}

func (d *desiredResourceListManager) DeleteResource(ctx context.Context, resource *vpcgw.Gateway) error {
	logf.FromContext(ctx).Info("Deleting Gateway", "gatewayName", resource.Name, "zone", resource.Zone)

	if err := d.Cloud().DeleteGateway(
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
	desired infrav1.PublicGateway,
) (*vpcgw.Gateway, error) {
	if desired.Type != "" && desired.Type != resource.Type {
		canUpgradeType, err := d.canUpgradeType(ctx, resource.Zone, resource.Type, desired.Type)
		if err != nil {
			return nil, err
		}

		if canUpgradeType {
			logf.FromContext(ctx).Info("Upgrading Gateway", "gatewayName", resource.Name, "zone", resource.Zone)
			return d.Cloud().UpgradeGateway(ctx, resource.Zone, resource.ID, desired.Type)
		}
	}

	return resource, nil
}

func (d *desiredResourceListManager) GetResourceZone(resource *vpcgw.Gateway) scw.Zone {
	return resource.Zone
}

func (d *desiredResourceListManager) GetResourceName(resource *vpcgw.Gateway) string {
	return resource.Name
}

func (d *desiredResourceListManager) GetDesiredZone(desired infrav1.PublicGateway) (scw.Zone, error) {
	return d.Cloud().GetZoneOrDefault(string(desired.Zone))
}

func (d *desiredResourceListManager) ShouldKeepResource(
	ctx context.Context,
	resource *vpcgw.Gateway,
	desired infrav1.PublicGateway,
) (bool, error) {
	// Gateway has no IPv4, remove it and recreate it.
	if resource.IPv4 == nil {
		return false, nil
	}

	if desired.Type != "" && desired.Type != resource.Type {
		canUpgradeType, err := d.canUpgradeType(ctx, resource.Zone, resource.Type, desired.Type)
		if err != nil {
			return false, err
		}

		if !canUpgradeType {
			return false, nil
		}
	}

	if desired.IP == "" && !slices.Contains(resource.Tags, capsManagedIPTag) {
		return false, nil
	}

	if desired.IP != "" && resource.IPv4.Address.String() != string(desired.IP) {
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
	desired infrav1.PublicGateway,
) (*vpcgw.Gateway, error) {
	var ipID *string
	var gwType string
	tags := d.ResourceTags()

	if desired.IP != "" {
		ip, err := d.Cloud().FindGatewayIP(ctx, zone, string(desired.IP))
		if err != nil {
			if client.IsNotFoundError(err) {
				return nil, fmt.Errorf("failed to find gateway ip: %w", err)
			}

			return nil, fmt.Errorf("failed to find gateway ip: %w", err)
		}

		ipID = &ip.ID
	} else {
		tags = append(tags, capsManagedIPTag)
	}

	if desired.Type != "" {
		gwType = desired.Type
	}

	logf.FromContext(ctx).Info("Creating Gateway", "gatewayName", name, "zone", zone)

	gateway, err := d.Cloud().CreateGateway(ctx, zone, name, gwType, tags, ipID)
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway: %w", err)
	}

	return gateway, nil
}

func (d *desiredResourceListManager) canUpgradeType(ctx context.Context, zone scw.Zone, current, desired string) (bool, error) {
	types, ok := d.gatewayTypesCache[zone]
	if !ok {
		var err error
		types, err = d.Cloud().ListGatewayTypes(ctx, zone)
		if err != nil {
			return false, err
		}

		d.gatewayTypesCache[zone] = types
	}

	return canUpgradeTypes(types, current, desired), nil
}

func canUpgradeTypes(types []string, current, desired string) bool {
	desiredIndex := slices.Index(types, desired)
	currentIndex := slices.Index(types, current)

	return currentIndex != -1 && desiredIndex > currentIndex
}
