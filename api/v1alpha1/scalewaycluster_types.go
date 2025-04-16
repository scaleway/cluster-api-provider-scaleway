package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScalewayClusterSpec defines the desired state of ScalewayCluster.
type ScalewayClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ScalewayCluster. Edit scalewaycluster_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// ScalewayClusterStatus defines the observed state of ScalewayCluster.
type ScalewayClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ScalewayCluster is the Schema for the scalewayclusters API.
type ScalewayCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScalewayClusterSpec   `json:"spec,omitempty"`
	Status ScalewayClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScalewayClusterList contains a list of ScalewayCluster.
type ScalewayClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScalewayCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScalewayCluster{}, &ScalewayClusterList{})
}
