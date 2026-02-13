package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck
)

const ManagedControlPlaneFinalizer = "scalewaycluster.infrastructure.cluster.x-k8s.io/smcp-protection"

// ScalewayManagedControlPlaneSpec defines the desired state of ScalewayManagedControlPlane
//
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.controlPlaneEndpoint) || has(self.controlPlaneEndpoint)", message="controlPlaneEndpoint is required once set"
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.clusterName) || has(self.clusterName) == has(oldSelf.clusterName)",message="clusterName cannot be removed once set"
// +kubebuilder:validation:XValidation:rule="has(self.cni) == has(oldSelf.cni)",message="cni cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.enablePrivateEndpoint) == has(oldSelf.enablePrivateEndpoint)",message="enablePrivateEndpoint cannot be added or removed"
type ScalewayManagedControlPlaneSpec struct {
	// ClusterName allows you to specify the name of the Scaleway managed cluster.
	// If you don't specify a name then a default name will be created
	// based on the namespace and name of the managed control plane.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:MaxLength:=100
	// +optional
	ClusterName *string `json:"clusterName,omitempty"`

	// Type of the cluster (e.g. kapsule, multicloud, etc.).
	// +kubebuilder:default="kapsule"
	Type string `json:"type"`

	// Version defines the desired Kubernetes version.
	// +kubebuilder:validation:MinLength:=2
	Version string `json:"version"`

	// CNI plugin running in the cluster.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:Enum=cilium;calico;kilo;none
	// +optional
	CNI *string `json:"cni,omitempty"`

	// AdditionalTags that will be added to the default tags.
	// +optional
	AdditionalTags []string `json:"additionalTags,omitempty"`

	// Autoscaler configuration of the cluster.
	// +optional
	Autoscaler *AutoscalerSpec `json:"autoscaler,omitempty"`

	// AutoUpgrade configuration of the cluster.
	// +optional
	AutoUpgrade *AutoUpgradeSpec `json:"autoUpgrade,omitempty"`

	// Feature gates to enable.
	// +optional
	FeatureGates []string `json:"featureGates,omitempty"`

	// Admission plugins to enable.
	// +optional
	AdmissionPlugins []string `json:"admissionPlugins,omitempty"`

	// OpenIDConnect defines the OpenID Connect configuration of the Kubernetes API server.
	OpenIDConnect *OpenIDConnectSpec `json:"openIDConnect,omitempty"`

	// APIServerCertSANs defines additional Subject Alternative Names for the
	// Kubernetes API server certificate.
	// +optional
	APIServerCertSANs []string `json:"apiServerCertSANs,omitempty"`

	// OnDelete configures the settings to apply when deleting the Scaleway managed cluster.
	// +optional
	OnDelete *OnDeleteSpec `json:"onDelete,omitempty"`

	// ACLSpec configures the ACLs of the managed cluster. If not set, ACLs
	// will be set to 0.0.0.0/0.
	// +optional
	ACL *ACLSpec `json:"acl,omitempty"`

	// EnablePrivateEndpoint defines whether the apiserver's internal address
	// is used as the cluster endpoint.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	EnablePrivateEndpoint *bool `json:"enablePrivateEndpoint,omitempty"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	ControlPlaneEndpoint clusterv1beta1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`
}

// OnDeleteSpec configures the settings to apply when deleting the Scaleway managed cluster.
type OnDeleteSpec struct {
	// WithAdditionalResources allows to also automatically delete all volumes
	// (including those with volume type "retain"), empty Private Networks and
	// Load Balancers whose names start with cluster ID.
	// +optional
	WithAdditionalResources *bool `json:"withAdditionalResources,omitempty"`
}

// AutoscalerSpec allows you to set (to an extent) your preferred autoscaler configuration,
// which is an implementation of the cluster-autoscaler (https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/).
type AutoscalerSpec struct {
	// Disable the cluster autoscaler.
	// +optional
	ScaleDownDisabled *bool `json:"scaleDownDisabled,omitempty"`

	// How long after scale up the scale down evaluation resumes.
	// +optional
	ScaleDownDelayAfterAdd *string `json:"scaleDownDelayAfterAdd,omitempty"`

	// Type of resource estimator to be used in scale up.
	// +kubebuilder:validation:Enum=binpacking
	// +optional
	Estimator *string `json:"estimator,omitempty"`

	// Type of node group expander to be used in scale up.
	// +kubebuilder:validation:Enum=random;most_pods;least_waste;priority;price
	// +optional
	Expander *string `json:"expander,omitempty"`

	// Ignore DaemonSet pods when calculating resource utilization for scaling down.
	// +optional
	IgnoreDaemonsetsUtilization *bool `json:"ignoreDaemonsetsUtilization,omitempty"`

	// Detect similar node groups and balance the number of nodes between them.
	// +optional
	BalanceSimilarNodeGroups *bool `json:"balanceSimilarNodeGroups,omitempty"`

	// Pods with priority below cutoff will be expendable. They can be killed without
	// any consideration during scale down and they won't cause scale up.
	// Pods with null priority (PodPriority disabled) are non expendable.
	ExpendablePodsPriorityCutoff *int32 `json:"expendablePodsPriorityCutoff,omitempty"`

	// How long a node should be unneeded before it is eligible to be scaled down.
	// +optional
	ScaleDownUnneededTime *string `json:"scaleDownUnneededTime,omitempty"`

	// Node utilization level, defined as a sum of requested resources divided
	// by capacity, below which a node can be considered for scale down.
	// +kubebuilder:validation:Format="float"
	// +optional
	ScaleDownUtilizationThreshold *string `json:"scaleDownUtilizationThreshold,omitempty"`

	// Maximum number of seconds the cluster autoscaler waits for pod termination
	// when trying to scale down a node.
	// +optional
	MaxGracefulTerminationSec *int32 `json:"maxGracefulTerminationSec,omitempty"`
}

// AutoUpgradeSpec allows to set a specific 2-hour time window in which the cluster
// can be automatically updated to the latest patch version.
type AutoUpgradeSpec struct {
	// Defines whether auto upgrade is enabled for the cluster.
	Enabled bool `json:"enabled"`

	// Maintenance window of the cluster auto upgrades.
	// +optional
	MaintenanceWindow *MaintenanceWindowSpec `json:"maintenanceWindow,omitempty"`
}

// MaintenanceWindowSpec defines the window of the cluster auto upgrades.
type MaintenanceWindowSpec struct {
	// Start time of the two-hour maintenance window.
	// +optional
	StartHour *int32 `json:"startHour,omitempty"`

	// Day of the week for the maintenance window.
	// +kubebuilder:validation:Enum=any;monday;tuesday;wednesday;thursday;friday;saturday;sunday
	// +optional
	Day *string `json:"day,omitempty"`
}

// OpenIDConnectSpec defines the OpenID Connect configuration of the Kubernetes API server.
type OpenIDConnectSpec struct {
	// URL of the provider which allows the API server to discover public signing keys.
	// Only URLs using the https:// scheme are accepted. This is typically the provider's
	// discovery URL without a path, for example "https://accounts.google.com" or "https://login.salesforce.com".
	IssuerURL string `json:"issuerURL"`

	// A client ID that all tokens must be issued for.
	ClientID string `json:"clientID"`

	// JWT claim to use as the user name. The default is "sub", which is expected
	// to be the end user's unique identifier. Admins can choose other claims,
	// such as email or name, depending on their provider. However, claims other
	// than email will be prefixed with the issuer URL to prevent name collision.
	// +optional
	UsernameClaim *string `json:"usernameClaim,omitempty"`

	// Prefix prepended to username claims to prevent name collision (such as "system:" users).
	// For example, the value "oidc:"" will create usernames like "oidc:jane.doe".
	// If this flag is not provided and "username_claim" is a value other than email,
	// the prefix defaults to "( Issuer URL )#" where "( Issuer URL )" is the value of "issuer_url".
	// The value "-" can be used to disable all prefixing.
	// +optional
	UsernamePrefix *string `json:"usernamePrefix,omitempty"`

	// JWT claim to use as the user's group.
	// +optional
	GroupsClaim []string `json:"groupsClaim,omitempty"`

	// Prefix prepended to group claims to prevent name collision (such as "system:" groups).
	// For example, the value "oidc:" will create group names like "oidc:engineering" and "oidc:infra".
	// +optional
	GroupsPrefix *string `json:"groupsPrefix,omitempty"`

	// Multiple key=value pairs describing a required claim in the ID token. If set,
	// the claims are verified to be present in the ID token with a matching value.
	// +optional
	RequiredClaim []string `json:"requiredClaim,omitempty"`
}

// ScalewayManagedControlPlaneStatus defines the observed state of ScalewayManagedControlPlane.
type ScalewayManagedControlPlaneStatus struct {
	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready,omitempty"`

	// Initialized is true when the control plane is available for initial contact.
	// This may occur before the control plane is fully ready.
	// +optional
	Initialized bool `json:"initialized,omitempty"`

	// ExternalManagedControlPlane is a bool that should be set to true if the
	// Node objects do not exist in the cluster.
	// +kubebuilder:default=true
	// +optional
	ExternalManagedControlPlane bool `json:"externalManagedControlPlane,omitempty"`

	// Version represents the version of the Scaleway managed control plane.
	// +optional
	Version *string `json:"version,omitempty"`
}

// ACLSpec configures the ACLs of the managed cluster.
type ACLSpec struct {
	// AllowedRanges allows to set a list of allowed public IP ranges that can access
	// the managed cluster. When empty, all IP ranges are DENIED. Make sure the nodes
	// of your management cluster can still access the cluster by allowing their IPs.
	// +kubebuilder:validation:MaxItems=30
	// +listType=set
	// +optional
	AllowedRanges []CIDR `json:"allowedRanges,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewaymanagedcontrolplanes,scope=Namespaced,categories=cluster-api,shortName=smcp
// +kubebuilder:subresource:status
// +kubebuilder:deprecatedversion
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this ScalewayManagedControlPlane belongs"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Ready is true when the managed cluster is fully provisioned"
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
	Spec ScalewayManagedControlPlaneSpec `json:"spec"`

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

func init() {
	objectTypes = append(objectTypes, &ScalewayManagedControlPlane{}, &ScalewayManagedControlPlaneList{})
}
