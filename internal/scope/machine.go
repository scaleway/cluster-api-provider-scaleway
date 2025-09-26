package scope

import (
	"context"
	"errors"
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

const (
	defaultRootVolumeSize = 20 * scw.GB
	defaultRootVolumeType = instance.VolumeVolumeTypeSbsVolume
)

var volumeTypeToInstanceVolumeType = map[string]instance.VolumeVolumeType{
	"local": instance.VolumeVolumeTypeLSSD,
	"block": instance.VolumeVolumeTypeSbsVolume,
}

type Machine struct {
	Client      client.Client
	patchHelper *patch.Helper

	*Cluster

	Machine         *clusterv1.Machine
	ScalewayMachine *infrav1.ScalewayMachine
}

// MachineParams contains mandatory params for creating the Machine scope.
type MachineParams struct {
	Client          client.Client
	ClusterScope    *Cluster
	Machine         *clusterv1.Machine
	ScalewayMachine *infrav1.ScalewayMachine
}

// NewMachine creates a new Machine scope.
func NewMachine(params *MachineParams) (*Machine, error) {
	helper, err := patch.NewHelper(params.ScalewayMachine, params.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch helper for ScalewayCluster: %w", err)
	}

	return &Machine{
		Client:          params.Client,
		patchHelper:     helper,
		Cluster:         params.ClusterScope,
		Machine:         params.Machine,
		ScalewayMachine: params.ScalewayMachine,
	}, nil
}

// PatchObject patches the ScalewayMachine object.
func (m *Machine) PatchObject(ctx context.Context) error {
	summaryConditions := []string{
		infrav1.ScalewayMachineInstanceReadyCondition,
	}

	if err := conditions.SetSummaryCondition(m.ScalewayMachine, m.ScalewayMachine, infrav1.ScalewayMachineReadyCondition, conditions.ForConditionTypes(summaryConditions)); err != nil {
		return err
	}

	return m.patchHelper.Patch(ctx, m.ScalewayMachine, patch.WithOwnedConditions{
		Conditions: append(summaryConditions, infrav1.ScalewayMachineReadyCondition),
	})
}

// Close closes the Machine scope by patching the ScalewayMachine object.
func (m *Machine) Close(ctx context.Context) error {
	return m.PatchObject(ctx)
}

// ResourceNameName returns the name that resources created for the machine should have.
func (m *Machine) ResourceName() string {
	return m.ScalewayMachine.Name
}

// ResourceTags returns the tags that resources created for the machine should have.
func (m *Machine) ResourceTags() []string {
	return append(m.Cluster.ResourceTags(), fmt.Sprintf("caps-scalewaymachine=%s", m.ScalewayMachine.Name))
}

// Zone returns the zone of the machine.
func (m *Machine) Zone() (scw.Zone, error) {
	return m.ScalewayClient.GetZoneOrDefault(m.Machine.Spec.FailureDomain)
}

// RootVolumeSize returns the size of the root volume for the machine.
func (m *Machine) RootVolumeSize() scw.Size {
	size := defaultRootVolumeSize

	if m.ScalewayMachine.Spec.RootVolume.Size != 0 {
		size = scw.Size(m.ScalewayMachine.Spec.RootVolume.Size) * scw.GB
	}

	return size
}

// RootVolumeType returns the type of the root volume for the machine.
func (m *Machine) RootVolumeType() (instance.VolumeVolumeType, error) {
	volumeType := defaultRootVolumeType

	if m.ScalewayMachine.Spec.RootVolume.Type != "" {
		volumeType = volumeTypeToInstanceVolumeType[m.ScalewayMachine.Spec.RootVolume.Type]
		if volumeType == "" {
			return "", fmt.Errorf("unknown volume type %s", m.ScalewayMachine.Spec.RootVolume.Type)
		}
	}

	return volumeType, nil
}

// RootVolumeIOPS returns the IOPS of the root volume for the machine.
// If not specified, it returns nil.
// Note: IOPS is only applicable for block volumes.
func (m *Machine) RootVolumeIOPS() *int64 {
	if m.ScalewayMachine.Spec.RootVolume.IOPS != 0 {
		return &m.ScalewayMachine.Spec.RootVolume.IOPS
	}

	return nil
}

// HasPublicIPv4 returns true if the machine should have a Public IPv4 address.
func (m *Machine) HasPublicIPv4() bool {
	// If the cluster has no Private Network, we must enable a Public IPv4 so that
	// the machine can access the Cluster loadbalancer.
	if !m.Cluster.HasPrivateNetwork() {
		return true
	}

	return ptr.Deref(m.ScalewayMachine.Spec.PublicNetwork.EnableIPv4, false)
}

// HasPublicIPv6 returns true if the machine should have a Public IPv6 address.
func (m *Machine) HasPublicIPv6() bool {
	return ptr.Deref(m.ScalewayMachine.Spec.PublicNetwork.EnableIPv6, false)
}

// SetProviderID sets the ProviderID of the ScalewayMachine if it is not already set.
func (m *Machine) SetProviderID(providerID string) {
	if m.ScalewayMachine.Spec.ProviderID == "" {
		m.ScalewayMachine.Spec.ProviderID = providerID
	}
}

// SetAddresses sets the addresses of the ScalewayMachine.
// It replaces the existing addresses with the provided ones.
func (m *Machine) SetAddresses(addresses []clusterv1.MachineAddress) {
	m.ScalewayMachine.Status.Addresses = addresses
}

// GetBootstrapData retrieves the bootstrap data from the secret specified in the ScalewayMachine.
// It returns an error if the secret is not found or if the value key is missing.
func (m *Machine) GetBootstrapData(ctx context.Context) ([]byte, error) {
	if m.Machine.Spec.Bootstrap.DataSecretName == nil {
		return nil, errors.New("missing bootstrap secret name in machine")
	}

	key := types.NamespacedName{Namespace: m.Machine.GetNamespace(), Name: *m.Machine.Spec.Bootstrap.DataSecretName}
	secret := &corev1.Secret{}
	if err := m.Client.Get(ctx, key, secret); err != nil {
		return nil, err
	}

	value, ok := secret.Data["value"]
	if !ok {
		return nil, errors.New("error retrieving bootstrap data: secret value key is missing")
	}

	return value, nil
}

// HasJoinedCluster returns true if the machine has joined the cluster.
// A machine is considered to have joined the cluster if it has a NodeRef with a non-empty name.
func (m *Machine) HasJoinedCluster() bool {
	return m.Machine.Status.NodeRef.IsDefined()
}

// IsControlPlane returns true if the machine is a control plane machine.
func (m *Machine) IsControlPlane() bool {
	return util.IsControlPlaneMachine(m.Machine)
}
