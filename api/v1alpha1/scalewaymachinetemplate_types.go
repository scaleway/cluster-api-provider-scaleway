package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScalewayMachineTemplateResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta metav1.ObjectMeta   `json:"metadata,omitempty"`
	Spec       ScalewayMachineSpec `json:"spec"`
}

// ScalewayMachineTemplateSpec defines the desired state of ScalewayMachineTemplate
type ScalewayMachineTemplateSpec struct {
<<<<<<< HEAD
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of ScalewayMachineTemplate. Edit scalewaymachinetemplate_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`
}

// ScalewayMachineTemplateStatus defines the observed state of ScalewayMachineTemplate.
type ScalewayMachineTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the ScalewayMachineTemplate resource.
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
	Template ScalewayMachineTemplateResource `json:"template"`
>>>>>>> tmp-original-13-02-26-16-17
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewaymachinetemplates,scope=Namespaced,categories=cluster-api,shortName=smt
// +kubebuilder:deprecatedversion

// ScalewayMachineTemplate is the Schema for the scalewaymachinetemplates API
type ScalewayMachineTemplate struct {
	metav1.TypeMeta `json:",inline"`

<<<<<<< HEAD
	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ScalewayMachineTemplate
	// +required
	Spec ScalewayMachineTemplateSpec `json:"spec"`

	// status defines the observed state of ScalewayMachineTemplate
	// +optional
	Status ScalewayMachineTemplateStatus `json:"status,omitzero"`
=======
	Spec ScalewayMachineTemplateSpec `json:"spec,omitempty"`
>>>>>>> tmp-original-13-02-26-16-17
}

// +kubebuilder:object:root=true

// ScalewayMachineTemplateList contains a list of ScalewayMachineTemplate
type ScalewayMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ScalewayMachineTemplate `json:"items"`
}

func init() {
	objectTypes = append(objectTypes, &ScalewayMachineTemplate{}, &ScalewayMachineTemplateList{})
}
