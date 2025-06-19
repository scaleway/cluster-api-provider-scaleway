package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const MachineFinalizer = "scalewaycluster.infrastructure.cluster.x-k8s.io/sm-protection"

// ScalewayMachineSpec defines the desired state of ScalewayMachine.
// +kubebuilder:validation:XValidation:rule="has(self.rootVolume) == has(oldSelf.rootVolume)",message="rootVolume cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.publicNetwork) == has(oldSelf.publicNetwork)",message="publicNetwork cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.placementGroup) == has(oldSelf.placementGroup)",message="placementGroup cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.securityGroup) == has(oldSelf.securityGroup)",message="securityGroup cannot be added or removed"
type ScalewayMachineSpec struct {
	// ProviderID must match the provider ID as seen on the node object corresponding to this machine.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	ProviderID *string `json:"providerID,omitempty"`

	// CommercialType of instance (e.g. PRO2-S).
	// +kubebuilder:default="PRO2-S"
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	CommercialType string `json:"commercialType"`

	// Image ID, Name or Label to use to create the instance.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Image ImageSpec `json:"image"`

	// RootVolume defines the characteristics of the system (root) volume.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	RootVolume *RootVolumeSpec `json:"rootVolume,omitempty"`

	// PublicNetwork allows attaching public IPs to the instance.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	PublicNetwork *PublicNetworkSpec `json:"publicNetwork,omitempty"`

	// PlacementGroup allows attaching a Placement Group to the instance.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	PlacementGroup *PlacementGroupSpec `json:"placementGroup,omitempty"`

	// SecurityGroup allows attaching a Security Group to the instance.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	SecurityGroup *SecurityGroupSpec `json:"securityGroup,omitempty"`
}

// RootVolumeSpec defines the characteristics of the system (root) volume.
type RootVolumeSpec struct {
	// Size of the root volume in GB. Defaults to 20 GB.
	// +kubebuilder:default=20
	// +kubebuilder:validation:Minimum=8
	// +optional
	Size *int64 `json:"size,omitempty"`

	// Type of the root volume. Can be local or block. Note that not all types
	// of instances support local volumes.
	// +kubebuilder:validation:Enum=local;block
	// +kubebuilder:default="block"
	// +optional
	Type *string `json:"type,omitempty"`

	// IOPS is the number of IOPS requested for the disk. This is only applicable for block volumes.
	// +optional
	IOPS *int64 `json:"iops,omitempty"`
}

// PublicNetworkSpec allows enabling the attachment of public IPs to the instance.
type PublicNetworkSpec struct {
	// EnableIPv4 defines whether server has IPv4 address enabled.
	// +optional
	EnableIPv4 *bool `json:"enableIPv4,omitempty"`
	// EnableIPv6 defines whether server has IPv6 addresses enabled.
	// +optional
	EnableIPv6 *bool `json:"enableIPv6,omitempty"`
}

// PlacementGroupSpec contains an ID or Name of an existing Placement Group.
// +kubebuilder:validation:XValidation:rule="(has(self.id) ? 1 : 0) + (has(self.name) ? 1 : 0) == 1",message="exactly one of id or name must be set"
type PlacementGroupSpec struct {
	// ID of the placement group.
	// +optional
	ID *string `json:"id,omitempty"`
	// Name of the placement group.
	// +optional
	Name *string `json:"name,omitempty"`
}

// SecurityGroupSpec contains an ID or Name of an existing Security Group.
// +kubebuilder:validation:XValidation:rule="(has(self.id) ? 1 : 0) + (has(self.name) ? 1 : 0) == 1",message="exactly one of id or name must be set"
type SecurityGroupSpec struct {
	// ID of the security group.
	// +optional
	ID *string `json:"id,omitempty"`
	// +optional
	// Name of the security group.
	Name *string `json:"name,omitempty"`
}

// ImageSpec contains an ID, Name or Label to use to create the instance.
// +kubebuilder:validation:XValidation:rule="(has(self.id) ? 1 : 0) + (has(self.name) ? 1 : 0) + (has(self.label) ? 1 : 0) == 1",message="exactly one of id, name or label must be set"
type ImageSpec struct {
	// ID of the image.
	ID *string `json:"id,omitempty"`
	// Name of the image.
	Name *string `json:"name,omitempty"`
	// Label of the image.
	Label *string `json:"label,omitempty"`
}

// ScalewayMachineStatus defines the observed state of ScalewayMachine.
type ScalewayMachineStatus struct {
	// Addresses contains the associated addresses for the machine.
	// +optional
	Addresses []clusterv1.MachineAddress `json:"addresses,omitempty"`

	// Ready denotes that the Scaleway machine infrastructure is fully provisioned.
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed. Please use conditions
	// to check the operational state of the infra machine.
	// +optional
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="CommercialType",type="string",JSONPath=".spec.commercialType",description="Instance commercial type"
// +kubebuilder:printcolumn:name="ProviderID",type="string",JSONPath=".spec.providerID",description="Node provider ID"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Indicates whether the Scaleway machine is ready"
// +kubebuilder:resource:path=scalewaymachines,scope=Namespaced,categories=cluster-api,shortName=sm
// +kubebuilder:storageversion

// ScalewayMachine is the Schema for the scalewaymachines API.
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() <= 63",message="name must be between 1 and 63 characters"
// +kubebuilder:validation:XValidation:rule="self.metadata.name.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",message="name must be a valid DNS label"
type ScalewayMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec   ScalewayMachineSpec   `json:"spec,omitempty"`
	Status ScalewayMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScalewayMachineList contains a list of ScalewayMachine.
type ScalewayMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScalewayMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScalewayMachine{}, &ScalewayMachineList{})
}
