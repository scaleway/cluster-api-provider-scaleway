package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScalewayClusterTemplateSpec defines the desired state of ScalewayClusterTemplate.
type ScalewayClusterTemplateSpec struct {
	Template ScalewayClusterTemplateResource `json:"template"`
}

type ScalewayClusterTemplateResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta metav1.ObjectMeta   `json:"metadata,omitempty"`
	Spec       ScalewayClusterSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewayclustertemplates,scope=Namespaced,categories=cluster-api,shortName=sct
// +kubebuilder:storageversion

// ScalewayClusterTemplate is the Schema for the scalewayclustertemplates API.
type ScalewayClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ScalewayClusterTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// ScalewayClusterTemplateList contains a list of ScalewayClusterTemplate.
type ScalewayClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScalewayClusterTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScalewayClusterTemplate{}, &ScalewayClusterTemplateList{})
}
