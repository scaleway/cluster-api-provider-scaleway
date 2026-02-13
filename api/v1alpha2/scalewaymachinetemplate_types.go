package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// ScalewayMachineTemplateSpec defines the desired state of ScalewayMachineTemplate
type ScalewayMachineTemplateSpec struct {
	// template is a ScalewayMachine template resource.
	// +required
	Template ScalewayMachineTemplateResource `json:"template,omitempty,omitzero"`
}

type ScalewayMachineTemplateResource struct {
	// metadata is a Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ScalewayMachine
	// +required
	Spec ScalewayMachineSpec `json:"spec,omitempty,omitzero"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewaymachinetemplates,scope=Namespaced,categories=cluster-api,shortName=smt
// +kubebuilder:storageversion

// ScalewayMachineTemplate is the Schema for the scalewaymachinetemplates API
type ScalewayMachineTemplate struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ScalewayMachineTemplate
	// +required
	Spec ScalewayMachineTemplateSpec `json:"spec,omitzero"`
}

// +kubebuilder:object:root=true

// ScalewayMachineTemplateList contains a list of ScalewayMachineTemplate
type ScalewayMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ScalewayMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScalewayMachineTemplate{}, &ScalewayMachineTemplateList{})
}
