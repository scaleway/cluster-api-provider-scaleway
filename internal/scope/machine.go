package scope

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	return m.patchHelper.Patch(ctx, m.ScalewayMachine)
}

// Close closes the Machine scope by patching the ScalewayMachine object.
func (m *Machine) Close(ctx context.Context) error {
	return m.PatchObject(ctx)
}

// ResourceNameName returns the name/prefix that resources created for the machine should have.
// It is possible to provide additional suffixes that will be appended to the name with a leading "-".
func (m *Machine) ResourceName(suffixes ...string) string {
	name := strings.Builder{}
	name.WriteString("caps")

	for _, suffix := range append([]string{m.ScalewayMachine.Name}, suffixes...) {
		name.WriteString("-")
		name.WriteString(suffix)
	}

	return truncateString(name.String())
}

// ResourceTags returns the tags that resources created for the machine should have.
// It is possible to provide additional tags that will be added to the default tags.
func (m *Machine) ResourceTags(additional ...string) []string {
	return slices.Concat(
		m.Cluster.ResourceTags(),
		[]string{fmt.Sprintf("caps-scalewaymachine=%s", m.ScalewayMachine.Name)},
		additional,
	)
}

// Zone returns the zone of the machine.
func (m *Machine) Zone() (scw.Zone, error) {
	return m.ScalewayClient.GetZoneOrDefault(m.Machine.Spec.FailureDomain)
}

// RootVolumeSize returns the size of the root volume for the machine.
func (m *Machine) RootVolumeSize() scw.Size {
	size := defaultRootVolumeSize

	if m.ScalewayMachine.Spec.RootVolume != nil &&
		m.ScalewayMachine.Spec.RootVolume.Size != nil {
		size = scw.Size(*m.ScalewayMachine.Spec.RootVolume.Size) * scw.GB
	}

	return size
}

// RootVolumeType returns the type of the root volume for the machine.
func (m *Machine) RootVolumeType() (instance.VolumeVolumeType, error) {
	volumeType := defaultRootVolumeType

	if m.ScalewayMachine.Spec.RootVolume != nil &&
		m.ScalewayMachine.Spec.RootVolume.Type != nil {
		volumeType = volumeTypeToInstanceVolumeType[*m.ScalewayMachine.Spec.RootVolume.Type]
		if volumeType == "" {
			return "", scaleway.WithTerminalError(fmt.Errorf("unknown volume type %s", *m.ScalewayMachine.Spec.RootVolume.Type))
		}
	}

	return volumeType, nil
}

// RootVolumeIOPS returns the IOPS of the root volume for the machine.
// If not specified, it returns nil.
// Note: IOPS is only applicable for block volumes.
func (m *Machine) RootVolumeIOPS() *int64 {
	if m.ScalewayMachine.Spec.RootVolume != nil {
		return m.ScalewayMachine.Spec.RootVolume.IOPS
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

	if m.ScalewayMachine.Spec.PublicNetwork != nil &&
		m.ScalewayMachine.Spec.PublicNetwork.EnableIPv4 != nil {
		return *m.ScalewayMachine.Spec.PublicNetwork.EnableIPv4
	}

	return false
}

// HasPublicIPv6 returns true if the machine should have a Public IPv6 address.
func (m *Machine) HasPublicIPv6() bool {
	if m.ScalewayMachine.Spec.PublicNetwork != nil &&
		m.ScalewayMachine.Spec.PublicNetwork.EnableIPv6 != nil {
		return *m.ScalewayMachine.Spec.PublicNetwork.EnableIPv6
	}

	return false
}

// SetProviderID sets the ProviderID of the ScalewayMachine if it is not already set.
func (m *Machine) SetProviderID(providerID string) {
	if m.ScalewayMachine.Spec.ProviderID == nil {
		m.ScalewayMachine.Spec.ProviderID = scw.StringPtr(providerID)
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
	return m.Machine.Status.NodeRef != nil && m.Machine.Status.NodeRef.Name != ""
}

// IsControlPlane returns true if the machine is a control plane machine.
func (m *Machine) IsControlPlane() bool {
	return util.IsControlPlaneMachine(m.Machine)
}
