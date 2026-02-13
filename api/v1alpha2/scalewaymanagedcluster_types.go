package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// ScalewayManagedClusterFinalizer is the finalizer that prevents deletion of a ScalewayManagedCluster.
const ScalewayManagedClusterFinalizer = "scalewaymanagedcluster.infrastructure.cluster.x-k8s.io/smc-protection"

// ScalewayManagedClusterReadyCondition reports if the ScalewayManagedCluster is ready.
const ScalewayManagedClusterReadyCondition = clusterv1.ReadyCondition

<<<<<<< HEAD
// ScalewayManagedClusterSpec defines the desired state of ScalewayManagedCluster
type ScalewayManagedClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of ScalewayManagedCluster. Edit scalewaymanagedcluster_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`
=======
// ScalewayManagedClusterSpec defines the desired state of ScalewayManagedCluster.
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.controlPlaneEndpoint) || has(self.controlPlaneEndpoint)", message="controlPlaneEndpoint is required once set"
// +kubebuilder:validation:XValidation:rule="(has(self.network) && has(self.network.privateNetwork)) == (has(oldSelf.network) && has(oldSelf.network.privateNetwork))",message="privateNetwork cannot be added or removed"
type ScalewayManagedClusterSpec struct {
	// region where the managed cluster will be created.
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Region ScalewayRegion `json:"region,omitempty"`

	// projectID in which the managed cluster will be created.
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ProjectID UUID `json:"projectID,omitempty"`

	// scalewaySecretName is the name of the secret that contains the Scaleway client parameters.
	// The following keys are required: SCW_ACCESS_KEY, SCW_SECRET_KEY.
	// The following key is optional: SCW_API_URL.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	ScalewaySecretName string `json:"scalewaySecretName,omitempty"`

	// network defines the network configuration of the managed cluster.
	// +optional
	Network ScalewayManagedClusterNetwork `json:"network,omitempty,omitzero"`

	// controlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint,omitempty,omitzero"`
}

// ScalewayManagedClusterNetwork defines the network configuration of a managed cluster.
// +kubebuilder:validation:MinProperties=1
type ScalewayManagedClusterNetwork struct {
	// privateNetwork allows attaching machines of the cluster to a Private Network.
	// +optional
	PrivateNetwork PrivateNetwork `json:"privateNetwork,omitempty,omitzero"`

	// publicGateways allows to manage Public Gateways that will be created and
	// attached to the Private Network of the cluster.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=6
	PublicGateways []PublicGateway `json:"publicGateways,omitempty"`
>>>>>>> tmp-original-13-02-26-16-17
}

// ScalewayManagedClusterStatus defines the observed state of ScalewayManagedCluster.
// +kubebuilder:validation:MinProperties=1
type ScalewayManagedClusterStatus struct {
<<<<<<< HEAD
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the ScalewayManagedCluster resource.
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
	// conditions represent the current state of the ScalewayManagedCluster resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// The status of each condition is one of True, False, or Unknown.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// initialization provides observations of the ScalewayManagedCluster initialization process.
	// NOTE: Fields in this struct are part of the Cluster API contract and are used to orchestrate initial Cluster provisioning.
	// +optional
	Initialization ScalewayManagedClusterInitializationStatus `json:"initialization,omitempty,omitzero"`

	// network contains information about currently provisioned network resources.
	// +optional
	Network ScalewayManagedClusterNetworkStatus `json:"network,omitempty,omitzero"`
}

// ScalewayManagedClusterInitializationStatus provides observations of the ScalewayManagedCluster initialization process.
// +kubebuilder:validation:MinProperties=1
type ScalewayManagedClusterInitializationStatus struct {
	// provisioned is true when the infrastructure provider reports that the Cluster's infrastructure is fully provisioned.
	// NOTE: this field is part of the Cluster API contract, and it is used to orchestrate initial Cluster provisioning.
	// +optional
	Provisioned *bool `json:"provisioned,omitempty"`
}

// ScalewayManagedClusterNetworkStatus contains information about currently provisioned network resources.
// +kubebuilder:validation:MinProperties=1
type ScalewayManagedClusterNetworkStatus struct {
	// privateNetworkID is the ID of the Private Network that is attached to the cluster.
	// +optional
	PrivateNetworkID UUID `json:"privateNetworkID,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewaymanagedclusters,scope=Namespaced,categories=cluster-api,shortName=smc
>>>>>>> tmp-original-13-02-26-16-17
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this ScalewayManagedCluster belongs"
// +kubebuilder:printcolumn:name="Provisioned",type="boolean",JSONPath=".status.initialization.provisioned",description="Provisioned is true when the cluster infrastructure is fully provisioned"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="ScalewayManagedCluster pass all readiness checks"
// +kubebuilder:printcolumn:name="Region",type="string",JSONPath=".spec.region",description="Region of the managed cluster"
// +kubebuilder:printcolumn:name="Host",type="string",JSONPath=".spec.controlPlaneEndpoint.host",description="Host of the control plane"
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".spec.controlPlaneEndpoint.port",description="Port of the control plane"

// ScalewayManagedCluster is the Schema for the scalewaymanagedclusters API
<<<<<<< HEAD
=======
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() <= 63",message="name must be between 1 and 63 characters"
// +kubebuilder:validation:XValidation:rule="self.metadata.name.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",message="name must be a valid DNS label"
>>>>>>> tmp-original-13-02-26-16-17
type ScalewayManagedCluster struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
<<<<<<< HEAD
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ScalewayManagedCluster
	// +required
	Spec ScalewayManagedClusterSpec `json:"spec"`

	// status defines the observed state of ScalewayManagedCluster
	// +optional
	Status ScalewayManagedClusterStatus `json:"status,omitzero"`
=======
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ScalewayManagedCluster
	// +required
	Spec ScalewayManagedClusterSpec `json:"spec,omitempty,omitzero"`

	// status defines the observed state of ScalewayManagedCluster
	// +optional
	Status ScalewayManagedClusterStatus `json:"status,omitempty,omitzero"`
>>>>>>> tmp-original-13-02-26-16-17
}

// +kubebuilder:object:root=true

// ScalewayManagedClusterList contains a list of ScalewayManagedCluster
type ScalewayManagedClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ScalewayManagedCluster `json:"items"`
}

// GetConditions returns the list of conditions for an ScalewayManagedCluster API object.
func (s *ScalewayManagedCluster) GetConditions() []metav1.Condition {
	return s.Status.Conditions
}

// SetConditions will set the given conditions on an ScalewayManagedCluster object.
func (s *ScalewayManagedCluster) SetConditions(conditions []metav1.Condition) {
	s.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&ScalewayManagedCluster{}, &ScalewayManagedClusterList{})
}
