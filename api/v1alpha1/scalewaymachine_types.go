package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScalewayMachineSpec defines the desired state of ScalewayMachine.
type ScalewayMachineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ScalewayMachine. Edit scalewaymachine_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// ScalewayMachineStatus defines the observed state of ScalewayMachine.
type ScalewayMachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=scalewaymachines,scope=Namespaced,categories=cluster-api,shortName=sm
// +kubebuilder:storageversion

// ScalewayMachine is the Schema for the scalewaymachines API.
type ScalewayMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

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
