package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ManagedMachinePoolFinalizer = "scalewaycluster.infrastructure.cluster.x-k8s.io/smmp-protection"

// ScalewayManagedMachinePoolSpec defines the desired state of ScalewayManagedMachinePool
//
// +kubebuilder:validation:XValidation:rule="has(self.placementGroupID) == has(oldSelf.placementGroupID)",message="placementGroupID cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.rootVolumeType) == has(oldSelf.rootVolumeType)",message="rootVolumeType cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.rootVolumeSizeGB) == has(oldSelf.rootVolumeSizeGB)",message="rootVolumeSizeGB cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.publicIPDisabled) == has(oldSelf.publicIPDisabled)",message="publicIPDisabled cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.securityGroupID) == has(oldSelf.securityGroupID)",message="securityGroupID cannot be added or removed"
type ScalewayManagedMachinePoolSpec struct {
	// NodeType is the type of Scaleway Instance wanted for the pool. Nodes with
	// insufficient memory are not eligible (DEV1-S, PLAY2-PICO, STARDUST).
	// "external" is a special node type used to provision instances from other
	// cloud providers in a Kosmos Cluster.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:MinLength:=2
	NodeType string `json:"nodeType"`

	// Zone in which the pool's nodes will be spawned.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:MinLength:=2
	Zone string `json:"zone"`

	// PlacementGroupID in which all the nodes of the pool will be created,
	// placement groups are limited to 20 instances.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	PlacementGroupID *string `json:"placementGroupID,omitempty"`

	// Scaling configures the scaling of the pool.
	// +optional
	Scaling *ScalingSpec `json:"scaling,omitempty"`

	// Autohealing defines whether the autohealing feature is enabled for the pool.
	// +optional
	Autohealing *bool `json:"autohealing,omitempty"`

	// AdditionalTags that will be added to the default tags.
	// +optional
	AdditionalTags []string `json:"additionalTags,omitempty"`

	// KubeletArgs defines Kubelet arguments to be used by this pool.
	// +optional
	KubeletArgs map[string]string `json:"kubeletArgs,omitempty"`

	// UpgradePolicy defines the pool's upgrade policy.
	// +optional
	UpgradePolicy *UpgradePolicySpec `json:"upgradePolicy,omitempty"`

	// RootVolumeType is the system volume disk type.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:Enum=l_ssd;sbs_5k;sbs_15k
	// +optional
	RootVolumeType *string `json:"rootVolumeType,omitempty"`

	// RootVolumeSizeGB is the size of the System volume disk size, in GB.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	RootVolumeSizeGB *int64 `json:"rootVolumeSizeGB,omitempty"`

	// PublicIPDisabled defines if the public IP should be removed from Nodes.
	// To use this feature, your Cluster must have an attached Private Network
	// set up with a Public Gateway.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	PublicIPDisabled *bool `json:"publicIPDisabled,omitempty"`

	// SecurityGroupID in which all the nodes of the pool will be created. If unset,
	// the pool will use default Kapsule security group in current zone.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	SecurityGroupID *string `json:"securityGroupID,omitempty"`

	// ProviderIDList are the provider IDs of instances in the
	// managed instance group corresponding to the nodegroup represented by this
	// machine pool
	// +optional
	ProviderIDList []string `json:"providerIDList,omitempty"`
}

// ScalingSpec defines the scaling parameters of the pool.
type ScalingSpec struct {
	// Autoscaling defines whether the autoscaling feature is enabled for the pool.
	// +optional
	Autoscaling *bool `json:"autoscaling,omitempty"`

	// MinSize defines the minimum size of the pool. Note that this field is only
	// used when autoscaling is enabled on the pool.
	// +optional
	MinSize *int32 `json:"minSize,omitempty"`

	// MaxSize defines the maximum size of the pool. Note that this field is only
	// used when autoscaling is enabled on the pool.
	// +optional
	MaxSize *int32 `json:"maxSize,omitempty"`
}

// UpgradePolicySpec defines the pool's upgrade policy.
type UpgradePolicySpec struct {
	// MaxUnavailable is the maximum number of available nodes during upgrades.
	// +kubebuilder:validation:Minimum=0
	// +optional
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`

	// MaxSurge is the maximum number of additional nodes that can be provisioned
	// during upgrades.
	// +kubebuilder:validation:Minimum=0
	// +optional
	MaxSurge *int32 `json:"maxSurge,omitempty"`
}

// ScalewayManagedMachinePoolStatus defines the observed state of ScalewayManagedMachinePool.
type ScalewayManagedMachinePoolStatus struct {
	// Ready denotes that the ScalewayManagedMachinePool has joined the cluster
	// +kubebuilder:default=false
	Ready bool `json:"ready"`
	// Replicas is the most recently observed number of replicas.
	// +optional
	Replicas int32 `json:"replicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewaymanagedmachinepools,scope=Namespaced,categories=cluster-api,shortName=smmp
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Replicas",type="string",JSONPath=".status.replicas"

// ScalewayManagedMachinePool is the Schema for the scalewaymanagedmachinepools API
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() <= 63",message="name must be between 1 and 63 characters"
// +kubebuilder:validation:XValidation:rule="self.metadata.name.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",message="name must be a valid DNS label"
type ScalewayManagedMachinePool struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ScalewayManagedMachinePool
	// +required
	Spec ScalewayManagedMachinePoolSpec `json:"spec"`

	// status defines the observed state of ScalewayManagedMachinePool
	// +optional
	Status ScalewayManagedMachinePoolStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// ScalewayManagedMachinePoolList contains a list of ScalewayManagedMachinePool
type ScalewayManagedMachinePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScalewayManagedMachinePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScalewayManagedMachinePool{}, &ScalewayManagedMachinePoolList{})
}
