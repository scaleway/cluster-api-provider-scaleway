package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// ScalewayClusterTemplateSpec defines the desired state of ScalewayClusterTemplate
type ScalewayClusterTemplateSpec struct {
	// template is a ScalewayCluster template resource.
	// +required
	Template ScalewayClusterTemplateResource `json:"template,omitempty,omitzero"`
}

type ScalewayClusterTemplateResource struct {
	// metadata is a Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ScalewayCluster
	// +required
	Spec ScalewayClusterSpec `json:"spec,omitempty,omitzero"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewayclustertemplates,scope=Namespaced,categories=cluster-api,shortName=sct
// +kubebuilder:storageversion

// ScalewayClusterTemplate is the Schema for the scalewayclustertemplates API
type ScalewayClusterTemplate struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ScalewayClusterTemplate
	// +required
	Spec ScalewayClusterTemplateSpec `json:"spec,omitzero"`
}

// +kubebuilder:object:root=true

// ScalewayClusterTemplateList contains a list of ScalewayClusterTemplate
type ScalewayClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ScalewayClusterTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScalewayClusterTemplate{}, &ScalewayClusterTemplateList{})
}
