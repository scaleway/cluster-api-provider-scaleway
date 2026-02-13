package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// ScalewayMachineFinalizer is the finalizer that prevents deletion of a ScalewayMachine.
const ScalewayMachineFinalizer = "scalewaymachine.infrastructure.cluster.x-k8s.io/sm-protection"

// ScalewayMachineReadyCondition reports if the ScalewayMachine is ready.
const ScalewayMachineReadyCondition = clusterv1.ReadyCondition

// ScalewayMachine's InstanceReady condition and corresponding reasons.
const (
	// ScalewayMachineInstanceReadyCondition indicates whether the Scaleway instance is ready.
	ScalewayMachineInstanceReadyCondition = "InstanceReady"

	// ScalewayMachineInstanceReadyReason surfaces when the Scaleway instance is ready.
	ScalewayMachineInstanceReadyReason = ReadyReason

	// ScalewayMachineInstanceReconciliationFailedReason surfaces when there is a failure in reconciling the Scaleway instance.
	ScalewayMachineInstanceReconciliationFailedReason = ReconciliationFailedReason
)

<<<<<<< HEAD
// ScalewayMachineSpec defines the desired state of ScalewayMachine
type ScalewayMachineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of ScalewayMachine. Edit scalewaymachine_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`
=======
// ScalewayMachineSpec defines the desired state of ScalewayMachine.
// +kubebuilder:validation:XValidation:rule="has(self.rootVolume) == has(oldSelf.rootVolume)",message="rootVolume cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.publicNetwork) == has(oldSelf.publicNetwork)",message="publicNetwork cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.placementGroup) == has(oldSelf.placementGroup)",message="placementGroup cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.securityGroup) == has(oldSelf.securityGroup)",message="securityGroup cannot be added or removed"
type ScalewayMachineSpec struct {
	// providerID must match the provider ID as seen on the node object corresponding to this machine.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=512
	ProviderID string `json:"providerID,omitempty"`

	// commercialType of instance (e.g. PRO2-S).
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=20
	CommercialType string `json:"commercialType,omitempty"`

	// image defines an image ID, Name or Label to use to create the instance.
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Image Image `json:"image,omitempty,omitzero"`

	// rootVolume defines the characteristics of the system (root) volume.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	RootVolume RootVolume `json:"rootVolume,omitempty,omitzero"`

	// publicNetwork allows attaching public IPs to the instance.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	PublicNetwork PublicNetwork `json:"publicNetwork,omitempty,omitzero"`

	// placementGroup allows attaching a Placement Group to the instance.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	PlacementGroup IDOrName `json:"placementGroup,omitempty,omitzero"`

	// securityGroup allows attaching a Security Group to the instance.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	SecurityGroup IDOrName `json:"securityGroup,omitempty,omitzero"`
}

// Image contains an ID, Name or Label to use to create the instance.
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type Image struct {
	IDOrName `json:",inline"`

	// label of the image (as defined in the marketplace).
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	Label string `json:"label,omitempty"`
}

// IDOrName contains an ID or Name of an existing Scaleway resource.
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type IDOrName struct {
	// id of the Scaleway resource.
	// +optional
	ID UUID `json:"id,omitempty"`

	// name of the Scaleway resource.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	Name string `json:"name,omitempty"`
}

// RootVolume defines the characteristics of the system (root) volume.
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:XValidation:rule="!has(self.iops) || has(self.type) && self.type == 'block'",message="iops can only be set for block volumes"
type RootVolume struct {
	// size of the root volume in GB. Defaults to 20 GB.
	// +optional
	// +kubebuilder:default=20
	// +kubebuilder:validation:Minimum=8
	// +kubebuilder:validation:Maximum=10000
	Size int64 `json:"size,omitempty"`

	// type of the root volume. Can be local or block. Note that not all types
	// of instances support local volumes.
	// +optional
	// +kubebuilder:default="block"
	// +kubebuilder:validation:Enum=local;block
	Type string `json:"type,omitempty"`

	// iops is the number of IOPS requested for the disk. This is only applicable for block volumes.
	// +optional
	// +kubebuilder:validation:Minimum=5000
	IOPS int64 `json:"iops,omitempty"`
}

// PublicNetwork allows enabling the attachment of public IPs to the instance.
// +kubebuilder:validation:MinProperties=1
type PublicNetwork struct {
	// enableIPv4 defines whether server should have an IPv4 created and attached.
	// +optional
	EnableIPv4 *bool `json:"enableIPv4,omitempty"`

	// enableIPv6 defines whether server should have an IPv6 created and attached.
	// +optional
	EnableIPv6 *bool `json:"enableIPv6,omitempty"`
>>>>>>> tmp-original-13-02-26-16-17
}

// ScalewayMachineStatus defines the observed state of ScalewayMachine.
// +kubebuilder:validation:MinProperties=1
type ScalewayMachineStatus struct {
<<<<<<< HEAD
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the ScalewayMachine resource.
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
=======
	// conditions represent the current state of the ScalewayMachine resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// The status of each condition is one of True, False, or Unknown.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// initialization provides observations of the ScalewayMachine initialization process.
	// NOTE: Fields in this struct are part of the Cluster API contract and are used to orchestrate initial Machine provisioning.
	// +optional
	Initialization ScalewayMachineInitializationStatus `json:"initialization,omitempty,omitzero"`

	// addresses contains the associated addresses for the machine.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=32
	Addresses []clusterv1.MachineAddress `json:"addresses,omitempty"`
}

// ScalewayMachineInitializationStatus provides observations of the ScalewayMachine initialization process.
// +kubebuilder:validation:MinProperties=1
type ScalewayMachineInitializationStatus struct {
	// provisioned is true when the infrastructure provider reports that the Machine's infrastructure is fully provisioned.
	// NOTE: this field is part of the Cluster API contract, and it is used to orchestrate initial Machine provisioning.
	// +optional
	Provisioned *bool `json:"provisioned,omitempty"`
>>>>>>> tmp-original-13-02-26-16-17
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewaymachines,scope=Namespaced,categories=cluster-api,shortName=sm
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="CommercialType",type="string",JSONPath=".spec.commercialType",description="Instance commercial type"
// +kubebuilder:printcolumn:name="ProviderID",type="string",JSONPath=".spec.providerID",description="Node provider ID"
// +kubebuilder:printcolumn:name="Provisioned",type="boolean",JSONPath=".status.initialization.provisioned",description="Provisioned is true when the machine infrastructure is fully provisioned"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="ScalewayMachine pass all readiness checks"

// ScalewayMachine is the Schema for the scalewaymachines API
<<<<<<< HEAD
=======
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() <= 63",message="name must be between 1 and 63 characters"
// +kubebuilder:validation:XValidation:rule="self.metadata.name.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",message="name must be a valid DNS label"
>>>>>>> tmp-original-13-02-26-16-17
type ScalewayMachine struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
<<<<<<< HEAD
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ScalewayMachine
	// +required
	Spec ScalewayMachineSpec `json:"spec"`

	// status defines the observed state of ScalewayMachine
	// +optional
	Status ScalewayMachineStatus `json:"status,omitzero"`
=======
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ScalewayMachine
	// +required
	Spec ScalewayMachineSpec `json:"spec,omitempty,omitzero"`

	// status defines the observed state of ScalewayMachine
	// +optional
	Status ScalewayMachineStatus `json:"status,omitempty,omitzero"`
>>>>>>> tmp-original-13-02-26-16-17
}

// +kubebuilder:object:root=true

// ScalewayMachineList contains a list of ScalewayMachine
type ScalewayMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ScalewayMachine `json:"items"`
}

// GetConditions returns the list of conditions for an ScalewayMachine API object.
func (s *ScalewayMachine) GetConditions() []metav1.Condition {
	return s.Status.Conditions
}

// SetConditions will set the given conditions on an ScalewayMachine object.
func (s *ScalewayMachine) SetConditions(conditions []metav1.Condition) {
	s.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&ScalewayMachine{}, &ScalewayMachineList{})
}
