package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

<<<<<<< HEAD
// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScalewayClusterTemplateSpec defines the desired state of ScalewayClusterTemplate
type ScalewayClusterTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of ScalewayClusterTemplate. Edit scalewayclustertemplate_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`
}

// ScalewayClusterTemplateStatus defines the observed state of ScalewayClusterTemplate.
type ScalewayClusterTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the ScalewayClusterTemplate resource.
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
>>>>>>> tmp-original-13-02-26-16-17
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewayclustertemplates,scope=Namespaced,categories=cluster-api,shortName=sct
// +kubebuilder:storageversion
<<<<<<< HEAD
// +kubebuilder:subresource:status
=======
>>>>>>> tmp-original-13-02-26-16-17

// ScalewayClusterTemplate is the Schema for the scalewayclustertemplates API
type ScalewayClusterTemplate struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
<<<<<<< HEAD
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ScalewayClusterTemplate
	// +required
	Spec ScalewayClusterTemplateSpec `json:"spec"`

	// status defines the observed state of ScalewayClusterTemplate
	// +optional
	Status ScalewayClusterTemplateStatus `json:"status,omitzero"`
=======
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ScalewayClusterTemplate
	// +required
	Spec ScalewayClusterTemplateSpec `json:"spec,omitempty,omitzero"`
>>>>>>> tmp-original-13-02-26-16-17
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
