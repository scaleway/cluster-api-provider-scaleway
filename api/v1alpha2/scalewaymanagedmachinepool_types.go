package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// ScalewayManagedMachinePoolFinalizer is the finalizer that prevents deletion of a ScalewayManagedMachinePool.
const ScalewayManagedMachinePoolFinalizer = "scalewaymanagedmachinepool.infrastructure.cluster.x-k8s.io/smmp-protection"

// ScalewayManagedMachinePoolReadyCondition reports if the ScalewayManagedMachinePool is ready.
const ScalewayManagedMachinePoolReadyCondition = clusterv1.ReadyCondition

// ScalewayManagedMachinePool's PoolReady condition and corresponding reasons.
const (
	// ScalewayManagedMachinePoolPoolReadyCondition indicates whether the Scaleway Kubernetes Pool is ready.
	ScalewayManagedMachinePoolPoolReadyCondition = "PoolReady"

	// ScalewayManagedMachinePoolPoolReadyReason surfaces when the Scaleway Kubernetes Pool is ready.
	ScalewayManagedMachinePoolPoolReadyReason = ReadyReason

	// ScalewayManagedMachinePoolPoolReconciliationFailedReason surfaces
	// when there is a failure in reconciling the Scaleway Kubernetes Pool.
	ScalewayManagedMachinePoolPoolReconciliationFailedReason = ReconciliationFailedReason

	// ScalewayManagedMachinePoolPoolTransientStatusReason surfaces when the
	// Scaleway Kubernetes Pool has a transient status.
	ScalewayManagedMachinePoolPoolTransientStatusReason = "TransientStatus"
)

<<<<<<< HEAD
// ScalewayManagedMachinePoolSpec defines the desired state of ScalewayManagedMachinePool
type ScalewayManagedMachinePoolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of ScalewayManagedMachinePool. Edit scalewaymanagedmachinepool_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`
=======
// ScalewayManagedMachinePoolSpec defines the desired state of ScalewayManagedMachinePool.
// +kubebuilder:validation:XValidation:rule="has(self.placementGroupID) == has(oldSelf.placementGroupID)",message="placementGroupID cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.rootVolumeType) == has(oldSelf.rootVolumeType)",message="rootVolumeType cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.rootVolumeSizeGB) == has(oldSelf.rootVolumeSizeGB)",message="rootVolumeSizeGB cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.publicIPDisabled) == has(oldSelf.publicIPDisabled)",message="publicIPDisabled cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.securityGroupID) == has(oldSelf.securityGroupID)",message="securityGroupID cannot be added or removed"
type ScalewayManagedMachinePoolSpec struct {
	// nodeType is the type of Scaleway Instance wanted for the pool. Nodes with
	// insufficient memory are not eligible (DEV1-S, PLAY2-PICO, STARDUST).
	// "external" is a special node type used to provision instances from other
	// cloud providers in a Kosmos Cluster.
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=30
	NodeType string `json:"nodeType,omitempty"`

	// zone in which the pool's nodes will be spawned.
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Zone ScalewayZone `json:"zone,omitempty"`

	// placementGroupID in which all the nodes of the pool will be created,
	// placement groups are limited to 20 instances.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	PlacementGroupID UUID `json:"placementGroupID,omitempty"`

	// scaling configures the scaling of the pool.
	// +optional
	Scaling Scaling `json:"scaling,omitempty,omitzero"`

	// autohealing defines whether the autohealing feature is enabled for the pool.
	// +optional
	Autohealing *bool `json:"autohealing,omitempty"`

	// additionalTags that will be added to the default tags.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=30
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=128
	AdditionalTags []string `json:"additionalTags,omitempty"`

	// kubeletArgs defines Kubelet arguments to be used by this pool.
	// +optional
	KubeletArgs map[string]string `json:"kubeletArgs,omitempty"`

	// upgradePolicy defines the pool's upgrade policy.
	// +optional
	UpgradePolicy UpgradePolicy `json:"upgradePolicy,omitempty,omitzero"`

	// rootVolumeType is the system volume disk type.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:Enum=l_ssd;sbs_5k;sbs_15k
	RootVolumeType string `json:"rootVolumeType,omitempty"`

	// rootVolumeSizeGB is the size of the System volume disk size, in GB.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:Minimum=20
	RootVolumeSizeGB int64 `json:"rootVolumeSizeGB,omitempty"`

	// publicIPDisabled defines if the public IP should be removed from Nodes.
	// To use this feature, your Cluster must have an attached Private Network
	// set up with a Public Gateway.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	PublicIPDisabled *bool `json:"publicIPDisabled,omitempty"`

	// securityGroupID in which all the nodes of the pool will be created. If unset,
	// the pool will use default Kapsule security group in current zone.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	SecurityGroupID UUID `json:"securityGroupID,omitempty"`

	// providerIDList are the identification IDs of machine instances provided by the provider.
	// This field must match the provider IDs as seen on the node objects corresponding to a machine pool's machine instances.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=10000
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=512
	ProviderIDList []string `json:"providerIDList,omitempty"`
}

// Scaling defines the scaling parameters of the pool.
// +kubebuilder:validation:MinProperties=1
type Scaling struct {
	// autoscaling defines whether the autoscaling feature is enabled for the pool.
	// +optional
	Autoscaling *bool `json:"autoscaling,omitempty"`

	// minSize defines the minimum size of the pool. Note that this field is only
	// used when autoscaling is enabled on the pool.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MinSize *int32 `json:"minSize,omitempty"`

	// maxSize defines the maximum size of the pool. Note that this field is only
	// used when autoscaling is enabled on the pool.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxSize *int32 `json:"maxSize,omitempty"`
}

// UpgradePolicy defines the pool's upgrade policy.
// +kubebuilder:validation:MinProperties=1
type UpgradePolicy struct {
	// maxUnavailable is the maximum number of available nodes during upgrades.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`

	// maxSurge is the maximum number of additional nodes that can be provisioned
	// during upgrades.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxSurge *int32 `json:"maxSurge,omitempty"`
>>>>>>> tmp-original-13-02-26-16-17
}

// ScalewayManagedMachinePoolStatus defines the observed state of ScalewayManagedMachinePool.
// +kubebuilder:validation:MinProperties=1
type ScalewayManagedMachinePoolStatus struct {
<<<<<<< HEAD
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the ScalewayManagedMachinePool resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
=======
	// conditions represent the current state of the ScalewayManagedMachinePool resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// The status of each condition is one of True, False, or Unknown.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ready is true when the provider resource is ready.
	// Deprecated: this field is kept for now as CAPI v1.11.3 still needs it.
	// The .initialization.provisioned field should be used instead.
	// +optional
	Ready *bool `json:"ready,omitempty"`

	// initialization provides observations of the ScalewayManagedMachinePool initialization process.
	// NOTE: Fields in this struct are part of the Cluster API contract and are used to orchestrate initial MachinePool provisioning.
	// +optional
	Initialization ScalewayManagedMachinePoolInitializationStatus `json:"initialization,omitempty,omitzero"`

	// replicas is the most recently observed number of replicas.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
}

// ScalewayManagedMachinePoolInitializationStatus provides observations of the ScalewayManagedMachinePool initialization process.
// +kubebuilder:validation:MinProperties=1
type ScalewayManagedMachinePoolInitializationStatus struct {
	// provisioned is true when the infrastructure provider reports that the MachinePool's infrastructure is fully provisioned.
	// +optional
	Provisioned *bool `json:"provisioned,omitempty"`
}

// +kubebuilder:object:root=true
>>>>>>> tmp-original-13-02-26-16-17
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=scalewaymanagedmachinepools,scope=Namespaced,categories=cluster-api,shortName=smmp
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Provisioned",type="boolean",JSONPath=".status.initialization.provisioned",description="Provisioned is true when the machinepool infrastructure is fully provisioned"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="ScalewayManagedMachinePool pass all readiness checks"
// +kubebuilder:printcolumn:name="Replicas",type="string",JSONPath=".status.replicas"

// ScalewayManagedMachinePool is the Schema for the scalewaymanagedmachinepools API
<<<<<<< HEAD
=======
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() <= 63",message="name must be between 1 and 63 characters"
// +kubebuilder:validation:XValidation:rule="self.metadata.name.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",message="name must be a valid DNS label"
>>>>>>> tmp-original-13-02-26-16-17
type ScalewayManagedMachinePool struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
<<<<<<< HEAD
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ScalewayManagedMachinePool
	// +required
	Spec ScalewayManagedMachinePoolSpec `json:"spec"`

	// status defines the observed state of ScalewayManagedMachinePool
	// +optional
	Status ScalewayManagedMachinePoolStatus `json:"status,omitzero"`
=======
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ScalewayManagedMachinePool
	// +required
	Spec ScalewayManagedMachinePoolSpec `json:"spec,omitempty,omitzero"`

	// status defines the observed state of ScalewayManagedMachinePool
	// +optional
	Status ScalewayManagedMachinePoolStatus `json:"status,omitempty,omitzero"`
>>>>>>> tmp-original-13-02-26-16-17
}

// +kubebuilder:object:root=true

// ScalewayManagedMachinePoolList contains a list of ScalewayManagedMachinePool
type ScalewayManagedMachinePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ScalewayManagedMachinePool `json:"items"`
}

// GetConditions returns the list of conditions for an ScalewayManagedMachinePool API object.
func (s *ScalewayManagedMachinePool) GetConditions() []metav1.Condition {
	return s.Status.Conditions
}

// SetConditions will set the given conditions on an ScalewayManagedMachinePool object.
func (s *ScalewayManagedMachinePool) SetConditions(conditions []metav1.Condition) {
	s.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&ScalewayManagedMachinePool{}, &ScalewayManagedMachinePoolList{})
}
