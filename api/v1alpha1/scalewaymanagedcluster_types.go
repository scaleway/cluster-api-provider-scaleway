package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck
)

const ManagedClusterFinalizer = "scalewaycluster.infrastructure.cluster.x-k8s.io/smc-protection"

// ScalewayManagedClusterSpec defines the desired state of ScalewayManagedCluster
<<<<<<< HEAD
type ScalewayManagedClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of ScalewayManagedCluster. Edit scalewaymanagedcluster_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`
=======
//
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.controlPlaneEndpoint) || has(self.controlPlaneEndpoint)", message="controlPlaneEndpoint is required once set"
// +kubebuilder:validation:XValidation:rule="(has(self.network) && has(self.network.privateNetwork)) == (has(oldSelf.network) && has(oldSelf.network.privateNetwork))",message="privateNetwork cannot be added or removed"
type ScalewayManagedClusterSpec struct {
	// Region where the managed cluster will be created.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:MinLength:=2
	Region string `json:"region"`

	// ProjectID in which the managed cluster will be created.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:MinLength:=2
	ProjectID string `json:"projectID"`

	// ScalewaySecretName is the name of the secret that contains the Scaleway client parameters.
	// The following keys are required: SCW_ACCESS_KEY, SCW_SECRET_KEY.
	// The following key is optional: SCW_API_URL.
	// +kubebuilder:validation:MinLength:=1
	ScalewaySecretName string `json:"scalewaySecretName"`

	// Network defines the network configuration of the managed cluster.
	// +optional
	Network *ManagedNetworkSpec `json:"network,omitempty"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	ControlPlaneEndpoint clusterv1beta1.APIEndpoint `json:"controlPlaneEndpoint,omitempty,omitzero"`
}

// ManagedNetworkSpec defines the network configuration of a managed cluster.
type ManagedNetworkSpec struct {
	// PrivateNetwork allows attaching machines of the cluster to a Private Network.
	// +kubebuilder:validation:XValidation:rule="has(self.vpcID) == has(oldSelf.vpcID)",message="vpcID cannot be added or removed"
	// +kubebuilder:validation:XValidation:rule="has(self.id) == has(oldSelf.id)",message="id cannot be added or removed"
	// +kubebuilder:validation:XValidation:rule="has(self.subnet) == has(oldSelf.subnet)",message="subnet cannot be added or removed"
	// +kubebuilder:validation:XValidation:rule="has(self.id) && !has(self.subnet) || !has(self.id)",message="subnet cannot be set when id is set"
	// +kubebuilder:validation:XValidation:rule="has(self.id) && !has(self.vpcID) || !has(self.id)",message="vpcID cannot be set when id is set"
	// +optional
	PrivateNetwork *PrivateNetworkParams `json:"privateNetwork,omitempty"`

	// PublicGateways allows to create Public Gateways that will be attached to the
	// Private Network of the cluster.
	// +kubebuilder:validation:MaxItems=6
	// +optional
	PublicGateways []PublicGatewaySpec `json:"publicGateways,omitempty"`
>>>>>>> tmp-original-13-02-26-16-17
}

// ScalewayManagedClusterStatus defines the observed state of ScalewayManagedCluster.
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
=======
	// Ready denotes that the Scaleway managed cluster infrastructure is fully provisioned.
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed.
	// +optional
	Ready bool `json:"ready"`

	// Network contains information about currently provisioned network resources.
	// +optional
	Network *ManagedNetworkStatus `json:"network,omitempty"`
}

// ManagedNetworkStatus contains information about currently provisioned network resources.
type ManagedNetworkStatus struct {
	// PrivateNetworkID is the ID of the Private Network that is attached to the cluster.
	// +optional
	PrivateNetworkID *string `json:"privateNetworkID,omitempty"`
>>>>>>> tmp-original-13-02-26-16-17
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewaymanagedclusters,scope=Namespaced,categories=cluster-api,shortName=smc
// +kubebuilder:subresource:status
// +kubebuilder:deprecatedversion
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this ScalewayManagedCluster belongs"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Ready is true when the managed cluster is fully provisioned"
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
=======
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`
>>>>>>> tmp-original-13-02-26-16-17

	// spec defines the desired state of ScalewayManagedCluster
	// +required
	Spec ScalewayManagedClusterSpec `json:"spec"`

	// status defines the observed state of ScalewayManagedCluster
	// +optional
<<<<<<< HEAD
	Status ScalewayManagedClusterStatus `json:"status,omitzero"`
=======
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

func init() {
	objectTypes = append(objectTypes, &ScalewayManagedCluster{}, &ScalewayManagedClusterList{})
}
