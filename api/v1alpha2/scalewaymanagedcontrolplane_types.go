package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// ScalewayManagedControlPlaneFinalizer is the finalizer that prevents deletion of a ScalewayManagedControlPlane.
const ScalewayManagedControlPlaneFinalizer = "scalewaymanagedcontrolplane.infrastructure.cluster.x-k8s.io/smcp-protection"

// ScalewayManagedControlPlaneReadyCondition reports if the ScalewayManagedControlPlane is ready.
const ScalewayManagedControlPlaneReadyCondition = clusterv1.ReadyCondition

// ScalewayManagedControlPlane's ClusterReady condition and corresponding reasons.
const (
	// ScalewayManagedControlPlaneClusterReadyCondition indicates whether
	// the Scaleway Kubernetes Cluster is ready.
	ScalewayManagedControlPlaneClusterReadyCondition = "ClusterReady"

	// ScalewayManagedControlPlaneClusterReadyReason surfaces when the Scaleway Kubernetes Cluster is ready.
	ScalewayManagedControlPlaneClusterReadyReason = ReadyReason

	// ScalewayManagedControlPlaneClusterReconciliationFailedReason surfaces
	// when there is a failure in reconciling the Scaleway Kubernetes Cluster.
	ScalewayManagedControlPlaneClusterReconciliationFailedReason = ReconciliationFailedReason

	// ScalewayManagedControlPlaneClusterTransientStatusReason surfaces when the
	// Scaleway Kubernetes Cluster has a transient status.
	ScalewayManagedControlPlaneClusterTransientStatusReason = "TransientStatus"
)

// ScalewayManagedControlPlaneSpec defines the desired state of ScalewayManagedControlPlane.
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.controlPlaneEndpoint) || has(self.controlPlaneEndpoint)", message="controlPlaneEndpoint is required once set"
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.clusterName) || has(self.clusterName) == has(oldSelf.clusterName)",message="clusterName cannot be removed once set"
// +kubebuilder:validation:XValidation:rule="has(self.cni) == has(oldSelf.cni)",message="cni cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.enablePrivateEndpoint) == has(oldSelf.enablePrivateEndpoint)",message="enablePrivateEndpoint cannot be added or removed"
type ScalewayManagedControlPlaneSpec struct {
	// clusterName allows you to specify the name of the Scaleway managed cluster.
	// If you don't specify a name then a default name will be created
	// based on the namespace and name of the managed control plane.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	ClusterName string `json:"clusterName,omitempty"`

	// type of the cluster (e.g. kapsule, multicloud, etc.).
	// +optional
	// +kubebuilder:default="kapsule"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=50
	Type string `json:"type,omitempty"`

	// version defines the desired Kubernetes version.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	Version string `json:"version,omitempty"`

	// cni plugin running in the cluster.
	// +optional
	// +kubebuilder:validation:Enum=cilium;cilium_native;calico;kilo;none
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	CNI string `json:"cni,omitempty"`

	// additionalTags that will be added to the default tags.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=30
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=128
	AdditionalTags []string `json:"additionalTags,omitempty"`

	// autoscaler configuration of the cluster.
	// +optional
	Autoscaler Autoscaler `json:"autoscaler,omitempty,omitzero"`

	// autoUpgrade configuration of the cluster.
	// +optional
	AutoUpgrade AutoUpgrade `json:"autoUpgrade,omitempty,omitzero"`

	// featureGates to enable.
	// +optional
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=128
	FeatureGates []string `json:"featureGates,omitempty"`

	// admissionPlugins to enable.
	// +optional
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=128
	AdmissionPlugins []string `json:"admissionPlugins,omitempty"`

	// openIDConnect defines the OpenID Connect configuration of the Kubernetes API server.
	// +optional
	OpenIDConnect OpenIDConnect `json:"openIDConnect,omitempty,omitzero"`

	// apiServerCertSANs defines additional Subject Alternative Names for the
	// Kubernetes API server certificate.
	// +optional
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=255
	APIServerCertSANs []string `json:"apiServerCertSANs,omitempty"`

	// onDelete configures the settings to apply when deleting the Scaleway managed cluster.
	// +optional
	OnDelete OnDelete `json:"onDelete,omitempty,omitzero"`

	// acl configures the ACLs of the managed cluster. If not set, ACLs will be set to [0.0.0.0/0].
	// +optional
	ACL *ACL `json:"acl,omitempty"`

	// enablePrivateEndpoint defines whether the apiserver's internal address
	// is used as the cluster endpoint.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	EnablePrivateEndpoint *bool `json:"enablePrivateEndpoint,omitempty"`

	// controlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint,omitempty,omitzero"`
}

// Autoscaler allows you to set (to an extent) your preferred autoscaler configuration,
// which is an implementation of the cluster-autoscaler (https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/).
// +kubebuilder:validation:MinProperties=1
type Autoscaler struct {
	// scaleDownDisabled allows to disable the cluster autoscaler.
	// +optional
	ScaleDownDisabled *bool `json:"scaleDownDisabled,omitempty"`

	// scaleDownDelayAfterAdd defines how long after scale up the scale down evaluation resumes.
	// +optional
	// +kubebuilder:validation:Format="duration"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=10
	ScaleDownDelayAfterAdd string `json:"scaleDownDelayAfterAdd,omitempty"`

	// estimator is the type of resource estimator to be used in scale up.
	// +optional
	// +kubebuilder:validation:Enum=binpacking
	Estimator string `json:"estimator,omitempty"`

	// expander is the type of node group expander to be used in scale up.
	// +optional
	// +kubebuilder:validation:Enum=random;most_pods;least_waste;priority;price
	Expander string `json:"expander,omitempty"`

	// ignoreDaemonsetsUtilization allows to ignore DaemonSet pods when calculating
	// resource utilization for scaling down.
	// +optional
	IgnoreDaemonsetsUtilization *bool `json:"ignoreDaemonsetsUtilization,omitempty"`

	// balanceSimilarNodeGroups allows to detect similar node groups and balance
	// the number of nodes between them.
	// +optional
	BalanceSimilarNodeGroups *bool `json:"balanceSimilarNodeGroups,omitempty"`

	// expendablePodsPriorityCutoff defines the priority threshold below which pods
	// are considered expendable. Pods with priority below cutoff will be expendable.
	// They can be killed without any consideration during scale down and they won't cause scale up.
	// Pods with null priority (PodPriority disabled) are non expendable.
	// +optional
	ExpendablePodsPriorityCutoff *int32 `json:"expendablePodsPriorityCutoff,omitempty"`

	// scaleDownUnneededTime defines how long a node should be unneeded before it
	// is eligible to be scaled down.
	// +optional
	// +kubebuilder:validation:Format="duration"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=10
	ScaleDownUnneededTime string `json:"scaleDownUnneededTime,omitempty"`

	// scaleDownUtilizationThreshold is the Node utilization level, defined as a
	// sum of requested resources divided by capacity, below which a node can be
	// considered for scale down.
	// +optional
	// +kubebuilder:validation:Format="float"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=10
	ScaleDownUtilizationThreshold string `json:"scaleDownUtilizationThreshold,omitempty"`

	// maxGracefulTerminationSec is the maximum number of seconds the cluster autoscaler
	// waits for pod termination when trying to scale down a node.
	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxGracefulTerminationSec int32 `json:"maxGracefulTerminationSec,omitempty"`
}

// AutoUpgrade allows to set a specific 2-hour time window in which the cluster
// can be automatically updated to the latest patch version.
type AutoUpgrade struct {
	// enabled defines whether auto upgrade is enabled for the cluster.
	// +required
	Enabled *bool `json:"enabled,omitempty"`

	// maintenanceWindow of the cluster auto upgrades.
	// +optional
	MaintenanceWindow MaintenanceWindow `json:"maintenanceWindow,omitempty,omitzero"`
}

// MaintenanceWindow defines the window of the cluster auto upgrades.
// +kubebuilder:validation:MinProperties=1
type MaintenanceWindow struct {
	// startHour is the start time of the two-hour maintenance window.
	// +optional
	// +kubebuilder:validation:Minimum=0
	StartHour *int32 `json:"startHour,omitempty"`

	// day of the week for the maintenance window.
	// +optional
	// +kubebuilder:validation:Enum=any;monday;tuesday;wednesday;thursday;friday;saturday;sunday
	Day string `json:"day,omitempty"`
}

// OpenIDConnect defines the OpenID Connect configuration of the Kubernetes API server.
type OpenIDConnect struct {
	// issuerURL of the provider which allows the API server to discover public signing keys.
	// Only URLs using the https:// scheme are accepted. This is typically the provider's
	// discovery URL without a path, for example "https://accounts.google.com" or "https://login.salesforce.com".
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	IssuerURL string `json:"issuerURL,omitempty"`

	// clientID is a client ID that all tokens must be issued for.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	ClientID string `json:"clientID,omitempty"`

	// usernameClaim is the JWT claim to use as the user name. The default is "sub",
	// which is expected to be the end user's unique identifier. Admins can choose other claims,
	// such as email or name, depending on their provider. However, claims other
	// than email will be prefixed with the issuer URL to prevent name collision.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	UsernameClaim string `json:"usernameClaim,omitempty"`

	// usernamePrefix is the prefix prepended to username claims to prevent name collision (such as "system:" users).
	// For example, the value "oidc:"" will create usernames like "oidc:jane.doe".
	// If this flag is not provided and "username_claim" is a value other than email,
	// the prefix defaults to "( Issuer URL )#" where "( Issuer URL )" is the value of "issuer_url".
	// The value "-" can be used to disable all prefixing.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	UsernamePrefix string `json:"usernamePrefix,omitempty"`

	// groupsClaim is the JWT claim to use as the user's group.
	// +optional
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=100
	GroupsClaim []string `json:"groupsClaim,omitempty"`

	// groupsPrefix is the prefix prepended to group claims to prevent name collision (such as "system:" groups).
	// For example, the value "oidc:" will create group names like "oidc:engineering" and "oidc:infra".
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	GroupsPrefix string `json:"groupsPrefix,omitempty"`

	// requiredClaim is multiple key=value pairs describing a required claim in the ID token.
	// If set, the claims are verified to be present in the ID token with a matching value.
	// +optional
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=100
	RequiredClaim []string `json:"requiredClaim,omitempty"`
}

// OnDelete configures the settings to apply when deleting the Scaleway managed cluster.
// +kubebuilder:validation:MinProperties=1
type OnDelete struct {
	// withAdditionalResources allows to also automatically delete all volumes
	// (including those with volume type "retain"), empty Private Networks and
	// Load Balancers whose names start with cluster ID.
	// +optional
	WithAdditionalResources *bool `json:"withAdditionalResources,omitempty"`
}

// ACL configures the ACLs of the managed cluster.
type ACL struct {
	// allowedRanges is a list of allowed public IP ranges that can access
	// the managed cluster. When empty, all IP ranges are DENIED. Make sure the nodes
	// of your management cluster can still access the cluster by allowing their IPs.
	// +kubebuilder:validation:MaxItems=30
	// +optional
	// +listType=set
	AllowedRanges []CIDR `json:"allowedRanges,omitempty"`
}

// ScalewayManagedControlPlaneStatus defines the observed state of ScalewayManagedControlPlane.
// +kubebuilder:validation:MinProperties=1
type ScalewayManagedControlPlaneStatus struct {
	// conditions represent the current state of the ScalewayManagedControlPlane resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// The status of each condition is one of True, False, or Unknown.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// version defines the desired Kubernetes version for the control plane.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	Version string `json:"version,omitempty"`

	// externalManagedControlPlane is a bool that should be set to true if the
	// Node objects do not exist in the cluster.
	// +optional
	// +kubebuilder:default=true
	ExternalManagedControlPlane *bool `json:"externalManagedControlPlane,omitempty"`

	// initialization provides observations of the ScalewayManagedControlPlane initialization process.
	// NOTE: Fields in this struct are part of the Cluster API contract and are used to orchestrate initial Cluster provisioning.
	// +optional
	Initialization ScalewayManagedControlPlaneInitializationStatus `json:"initialization,omitempty,omitzero"`
}

// ScalewayManagedControlPlaneInitializationStatus provides observations of the ScalewayManagedControlPlane initialization process.
// +kubebuilder:validation:MinProperties=1
type ScalewayManagedControlPlaneInitializationStatus struct {
	// controlPlaneInitialized is true when the control plane provider reports that the Kubernetes control plane is initialized;
	// usually a control plane is considered initialized when it can accept requests, no matter if this happens before
	// the control plane is fully provisioned or not.
	// NOTE: this field is part of the Cluster API contract, and it is used to orchestrate initial Cluster provisioning.
	// +optional
	ControlPlaneInitialized *bool `json:"controlPlaneInitialized,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=scalewaymanagedcontrolplanes,scope=Namespaced,categories=cluster-api,shortName=smcp
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this ScalewayManagedControlPlane belongs"
// +kubebuilder:printcolumn:name="Initialized",type=boolean,JSONPath=".status.initialization.controlPlaneInitialized",description="This denotes whether or not the control plane can accept requests"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="ScalewayManagedControlPlane pass all readiness checks"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version",description="The Kubernetes version of the Scaleway control plane"
// +kubebuilder:printcolumn:name="Host",type="string",JSONPath=".spec.controlPlaneEndpoint.host",description="Host of the control plane"
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".spec.controlPlaneEndpoint.port",description="Port of the control plane"

// ScalewayManagedControlPlane is the Schema for the scalewaymanagedcontrolplanes API
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() <= 63",message="name must be between 1 and 63 characters"
// +kubebuilder:validation:XValidation:rule="self.metadata.name.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",message="name must be a valid DNS label"
type ScalewayManagedControlPlane struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ScalewayManagedControlPlane
	// +required
	Spec ScalewayManagedControlPlaneSpec `json:"spec,omitempty,omitzero"`

	// status defines the observed state of ScalewayManagedControlPlane
	// +optional
	Status ScalewayManagedControlPlaneStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// ScalewayManagedControlPlaneList contains a list of ScalewayManagedControlPlane
type ScalewayManagedControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScalewayManagedControlPlane `json:"items"`
}

// GetConditions returns the list of conditions for an ScalewayManagedControlPlane API object.
func (s *ScalewayManagedControlPlane) GetConditions() []metav1.Condition {
	return s.Status.Conditions
}

// SetConditions will set the given conditions on an ScalewayManagedControlPlane object.
func (s *ScalewayManagedControlPlane) SetConditions(conditions []metav1.Condition) {
	s.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&ScalewayManagedControlPlane{}, &ScalewayManagedControlPlaneList{})
}
