package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// ScalewayClusterFinalizer is the finalizer that prevents deletion of a ScalewayCluster.
const ScalewayClusterFinalizer = "scalewaycluster.infrastructure.cluster.x-k8s.io/sc-protection"

// ScalewayClusterReadyCondition reports if the ScalewayCluster is ready.
const ScalewayClusterReadyCondition = clusterv1.ReadyCondition

// ScalewayCluster's DomainReady condition and corresponding reasons.
const (
	// DomainReadyCondition indicates whether the domain for the control plane endpoint is ready.
	ScalewayClusterDomainReadyCondition = "DomainReady"

	// ScalewayClusterNoDomainReason surfaces when no domain is defined in the spec.
	// In this case, the condition is set to True as there is nothing to configure.
	ScalewayClusterNoDomainReason = "NoDomain"

	// ScalewayClusterDomainReconciliationFailedReason surfaces when the domain reconciliation failed.
	ScalewayClusterDomainReconciliationFailedReason = ReconciliationFailedReason

	// ScalewayClusterDomainZoneConfiguredReason surfaces when the domain zone has been successfully configured.
	ScalewayClusterDomainZoneConfiguredReason = "ZoneConfigured"
)

// ScalewayCluster's LoadBalancersReady condition and corresponding reasons.
const (
	// LoadBalancersReadyCondition indicates whether the load balancers for the control plane endpoint are ready.
	ScalewayClusterLoadBalancersReadyCondition = "LoadBalancersReady"

	// ScalewayClusterLoadBalancersReadyReason surfaces when the load balancers are provisioned and ready.
	ScalewayClusterLoadBalancersReadyReason = ReadyReason

	// ScalewayClusterLoadBalancersNotReadyReason surfaces when one or multiple load balancers are not ready.
	ScalewayClusterLoadBalancersNotReadyReason = NotReadyReason

	// ScalewayClusterLoadBalancersInternalErrorReason surfaces when an unexpected error has occurred.
	ScalewayClusterLoadBalancersInternalErrorReason = InternalErrorReason

	// ScalewayClusterMainLoadBalancerReconciliationFailedReason surfaces when the main load balancer reconciliation failed.
	ScalewayClusterMainLoadBalancerReconciliationFailedReason = "MainLoadBalancerReconciliationFailed"

	// ScalewayClusterExtraLoadBalancersReconciliationFailedReason surfaces when the extra load balancers reconciliation failed.
	ScalewayClusterExtraLoadBalancersReconciliationFailedReason = "ExtraLoadBalancersReconciliationFailed"

	// ScalewayClusterLoadBalancerPrivateNetworkAttachmentFailedReason when the attachment of the load balancers to the private network failed.
	ScalewayClusterLoadBalancerPrivateNetworkAttachmentFailedReason = PrivateNetworkAttachmentFailedReason

	// ScalewayClusterBackendReconciliationFailedReason surfaces when the backend reconciliation failed.
	ScalewayClusterBackendReconciliationFailedReason = "BackendReconciliationFailed"

	// ScalewayClusterFrontendReconciliationFailedReason surfaces when the frontend reconciliation failed.
	ScalewayClusterFrontendReconciliationFailedReason = "FrontendReconciliationFailed"

	// ScalewayClusterLoadBalancerACLReconciliationFailedReason surfaces when the load balancer ACL reconciliation failed.
	ScalewayClusterLoadBalancerACLReconciliationFailedReason = "LoadBalancerACLReconciliationFailed"
)

<<<<<<< HEAD
// ScalewayClusterSpec defines the desired state of ScalewayCluster
type ScalewayClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of ScalewayCluster. Edit scalewaycluster_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`
=======
// ScalewayClusterSpec defines the desired state of ScalewayCluster.
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.controlPlaneEndpoint) || has(self.controlPlaneEndpoint)", message="controlPlaneEndpoint is required once set"
// +kubebuilder:validation:XValidation:rule="(has(self.network) && has(self.network.controlPlaneDNS)) == (has(oldSelf.network) && has(oldSelf.network.controlPlaneDNS))",message="controlPlaneDNS cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="(has(self.network) && has(self.network.privateNetwork)) == (has(oldSelf.network) && has(oldSelf.network.privateNetwork))",message="privateNetwork cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="(has(self.network) && has(self.network.controlPlaneLoadBalancer) && has(self.network.controlPlaneLoadBalancer.private)) == (has(oldSelf.network) && has(oldSelf.network.controlPlaneLoadBalancer) && has(oldSelf.network.controlPlaneLoadBalancer.private))",message="private cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="(has(self.network) && has(self.network.controlPlaneLoadBalancer) && has(self.network.controlPlaneLoadBalancer.ip)) == (has(oldSelf.network) && has(oldSelf.network.controlPlaneLoadBalancer) && has(oldSelf.network.controlPlaneLoadBalancer.ip))",message="ip cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="(has(self.network) && has(self.network.controlPlaneLoadBalancer) && has(self.network.controlPlaneLoadBalancer.zone)) == (has(oldSelf.network) && has(oldSelf.network.controlPlaneLoadBalancer) && has(oldSelf.network.controlPlaneLoadBalancer.zone))",message="zone cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="(has(self.network) && has(self.network.controlPlaneLoadBalancer) && has(self.network.controlPlaneLoadBalancer.privateIP)) == (has(oldSelf.network) && has(oldSelf.network.controlPlaneLoadBalancer) && has(oldSelf.network.controlPlaneLoadBalancer.privateIP))",message="privateIP cannot be added or removed"
type ScalewayClusterSpec struct {
	// projectID is the ID of a Scaleway project where the cluster will be created.
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ProjectID UUID `json:"projectID,omitempty"`

	// region represents the region where the cluster will be hosted.
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Region ScalewayRegion `json:"region,omitempty"`

	// scalewaySecretName is the name of the secret that contains the Scaleway client parameters.
	// The following keys are required: SCW_ACCESS_KEY, SCW_SECRET_KEY.
	// The following key is optional: SCW_API_URL.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	ScalewaySecretName string `json:"scalewaySecretName,omitempty"`

	// failureDomains is a list of failure domains where the control-plane nodes will be created.
	// Failure domains correspond to Scaleway zones inside the cluster region (e.g. fr-par-1).
	// +optional
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=3
	FailureDomains []ScalewayZone `json:"failureDomains,omitempty"`

	// network contains network related options for the cluster.
	// +optional
	Network ScalewayClusterNetwork `json:"network,omitempty,omitzero"`

	// controlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint,omitempty,omitzero"`
}

// ScalewayClusterNetwork defines network settings for a ScalewayCluster.
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:XValidation:rule="!has(self.controlPlaneExtraLoadBalancers) || has(self.controlPlaneDNS)",message="controlPlaneDNS is required when controlPlaneExtraLoadBalancers is set"
// +kubebuilder:validation:XValidation:rule="!has(self.publicGateways) || has(self.privateNetwork) && self.privateNetwork.enabled",message="privateNetwork is required when publicGateways is set"
// +kubebuilder:validation:XValidation:rule="!has(self.controlPlaneLoadBalancer) || !has(self.controlPlaneLoadBalancer.private) || !self.controlPlaneLoadBalancer.private || has(self.privateNetwork) && self.privateNetwork.enabled",message="privateNetwork is required when private LoadBalancer is enabled"
// +kubebuilder:validation:XValidation:rule="!has(self.controlPlaneDNS) || has(self.controlPlaneDNS) && has(self.controlPlaneDNS.domain) || has(self.controlPlaneDNS) && !has(self.controlPlaneDNS.domain) && has(self.controlPlaneLoadBalancer) && has(self.controlPlaneLoadBalancer.private) && self.controlPlaneLoadBalancer.private",message=".controlPlaneDNS.domain must be set unless control plane load balancer is private"
type ScalewayClusterNetwork struct {
	// controlPlaneLoadBalancer defines settings for the load balancer of the control plane.
	// +optional
	ControlPlaneLoadBalancer ControlPlaneLoadBalancer `json:"controlPlaneLoadBalancer,omitempty,omitzero"`

	// controlPlaneExtraLoadBalancers allows configuring additional load balancers.
	// Because Scaleway load balancers are currently zonal resources, you may set
	// up to 3 additional load balancers for achieving regional redundancy. It is
	// mandatory to set the controlPlaneDNS field when you do so.
	// NOTE: This may be removed in the future, when Scaleway supports regional LoadBalancers.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=3
	ControlPlaneExtraLoadBalancers []LoadBalancer `json:"controlPlaneExtraLoadBalancers,omitempty"`

	// controlPlaneDNS allows configuring a Scaleway Domain DNS Zone.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ControlPlaneDNS ControlPlaneDNS `json:"controlPlaneDNS,omitempty,omitzero"`

	// privateNetwork allows attaching machines of the cluster to a Private Network.
	// +optional
	PrivateNetwork PrivateNetworkSpec `json:"privateNetwork,omitempty,omitzero"`

	// publicGateways allows to manage Public Gateways that will be created and
	// attached to the Private Network of the cluster.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=6
	PublicGateways []PublicGateway `json:"publicGateways,omitempty"`
}

// LoadBalancer defines load balancer parameters.
// +kubebuilder:validation:MinProperties=1
type LoadBalancer struct {
	// zone where to create the load balancer. Must be in the same region as the
	// cluster. Defaults to the first zone of the region.
	// +optional
	Zone ScalewayZone `json:"zone,omitempty"`

	// type is the load balancer commercial offer type.
	// +optional
	// +kubebuilder:default="LB-S"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=10
	Type string `json:"type,omitempty"`

	// ip is an existing public IPv4 to use when creating a load balancer.
	// +optional
	IP IPv4 `json:"ip,omitempty"`

	// privateIP is an existing private IPv4 inside the Private Network to use
	// when attaching a load balancer to a Private Network. It must be pre-booked
	// inside the Scaleway IPAM.
	// +optional
	PrivateIP IPv4 `json:"privateIP,omitempty"`
}

// ControlPlaneLoadBalancer defines control plane load balancer settings.
// +kubebuilder:validation:MinProperties=1
type ControlPlaneLoadBalancer struct {
	// +kubebuilder:validation:XValidation:rule="!has(oldSelf.ip) || self.ip == oldSelf.ip",message="ip is immutable"
	// +kubebuilder:validation:XValidation:rule="!has(oldSelf.zone) || self.zone == oldSelf.zone",message="zone is immutable"
	// +kubebuilder:validation:XValidation:rule="!has(oldSelf.privateIP) || self.privateIP == oldSelf.privateIP",message="privateIP is immutable"
	LoadBalancer `json:",inline"`

	// allowedRanges allows to set a list of allowed IP ranges that can access
	// the cluster through the load balancer. When unset, all IP ranges are allowed.
	// To allow the cluster to work properly, public IPs of nodes and Public
	// Gateways will automatically be allowed. However, if this field is set,
	// you MUST manually allow IPs of the nodes of your management cluster.
	// +optional
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=30
	AllowedRanges []CIDR `json:"allowedRanges,omitempty"`

	// private disables the creation of a public IP on the load balancers when it's set to true.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Private *bool `json:"private,omitempty"`
}

// ControlPlaneDNS defines the DNS configuration of the control plane endpoint.
type ControlPlaneDNS struct {
	// domain is the DNS Zone that this record should live in. It must be pre-existing in your Scaleway account.
	// The format must be a string that conforms to the definition of a subdomain in DNS (RFC 1123).
	// This is optional if the control plane load balancer is private.
	// +optional
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Domain string `json:"domain,omitempty"`

	// name is the DNS short name of the record (non-FQDN). The format must consist of
	// alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.
	// +required
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name,omitempty"`
}

// IsDefined returns true if the ControlPlaneDNS is set.
func (c *ControlPlaneDNS) IsDefined() bool {
	if c == nil {
		return false
	}
	return c.Name != "" || c.Domain != ""
}

// PrivateNetworkSpec defines Private Network settings for the cluster.
type PrivateNetworkSpec struct {
	PrivateNetwork `json:",inline"`

	// enabled allows to automatically attach machines to a Private Network when it's set to true.
	// The Private Network is automatically created if no existing Private
	// Network ID is provided.
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Enabled *bool `json:"enabled,omitempty"`
>>>>>>> tmp-original-13-02-26-16-17
}

// ScalewayClusterStatus defines the observed state of ScalewayCluster.
// +kubebuilder:validation:MinProperties=1
type ScalewayClusterStatus struct {
<<<<<<< HEAD
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the ScalewayCluster resource.
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
	// conditions represent the current state of the ScalewayCluster resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// The status of each condition is one of True, False, or Unknown.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// failureDomains is a list of failure domain objects synced from the infrastructure provider.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=100
	FailureDomains []clusterv1.FailureDomain `json:"failureDomains,omitempty"`

	// initialization provides observations of the ScalewayCluster initialization process.
	// NOTE: Fields in this struct are part of the Cluster API contract and are used to orchestrate initial Cluster provisioning.
	// +optional
	Initialization ScalewayClusterInitializationStatus `json:"initialization,omitempty,omitzero"`

	// network contains information about network resources of the cluster.
	// +optional
	Network ScalewayClusterNetworkStatus `json:"network,omitempty,omitzero"`
}

// ScalewayClusterInitializationStatus provides observations of the ScalewayCluster initialization process.
// +kubebuilder:validation:MinProperties=1
type ScalewayClusterInitializationStatus struct {
	// provisioned is true when the infrastructure provider reports that the Cluster's infrastructure is fully provisioned.
	// NOTE: this field is part of the Cluster API contract, and it is used to orchestrate initial Cluster provisioning.
	// +optional
	Provisioned *bool `json:"provisioned,omitempty"`
}

// ScalewayClusterNetworkStatus contains information about network resources of the cluster.
// +kubebuilder:validation:MinProperties=1
type ScalewayClusterNetworkStatus struct {
	// vpcID is set if the cluster has an associated Private Network.
	// +optional
	VPCID UUID `json:"vpcID,omitempty"`

	// privateNetworkID is set if the cluster has an associated Private Network.
	// +optional
	PrivateNetworkID UUID `json:"privateNetworkID,omitempty"`

	// publicGatewayIDs is a list of Public Gateway IDs.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	PublicGatewayIDs []UUID `json:"publicGatewayIDs,omitempty"`

	// loadBalancerIP is the public IP of the cluster control-plane.
	// +optional
	LoadBalancerIP IPv4 `json:"loadBalancerIP,omitempty"`

	// extraLoadBalancerIPs is a list of IPs of the extra loadbalancers.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	ExtraLoadBalancerIPs []IPv4 `json:"extraLoadBalancerIPs,omitempty"`
>>>>>>> tmp-original-13-02-26-16-17
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=scalewayclusters,scope=Namespaced,categories=cluster-api,shortName=sc
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Host",type="string",JSONPath=".spec.controlPlaneEndpoint.host",description="Host of the control plane"
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".spec.controlPlaneEndpoint.port",description="Port of the control plane"
// +kubebuilder:printcolumn:name="Region",type="string",JSONPath=".spec.region",description="Region of the cluster"
// +kubebuilder:printcolumn:name="Provisioned",type="boolean",JSONPath=".status.initialization.provisioned",description="Provisioned is true when the cluster infrastructure is fully provisioned"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="ScalewayCluster pass all readiness checks"

// ScalewayCluster is the Schema for the scalewayclusters API
<<<<<<< HEAD
=======
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() <= 63",message="name must be between 1 and 63 characters"
// +kubebuilder:validation:XValidation:rule="self.metadata.name.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?$')",message="name must be a valid DNS label"
>>>>>>> tmp-original-13-02-26-16-17
type ScalewayCluster struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
<<<<<<< HEAD
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ScalewayCluster
	// +required
	Spec ScalewayClusterSpec `json:"spec"`

	// status defines the observed state of ScalewayCluster
	// +optional
	Status ScalewayClusterStatus `json:"status,omitzero"`
=======
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ScalewayCluster
	// +required
	Spec ScalewayClusterSpec `json:"spec,omitempty,omitzero"`

	// status defines the observed state of ScalewayCluster
	// +optional
	Status ScalewayClusterStatus `json:"status,omitempty,omitzero"`
>>>>>>> tmp-original-13-02-26-16-17
}

// +kubebuilder:object:root=true

// ScalewayClusterList contains a list of ScalewayCluster
type ScalewayClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ScalewayCluster `json:"items"`
}

// GetConditions returns the list of conditions for an ScalewayCluster API object.
func (s *ScalewayCluster) GetConditions() []metav1.Condition {
	return s.Status.Conditions
}

// SetConditions will set the given conditions on an ScalewayCluster object.
func (s *ScalewayCluster) SetConditions(conditions []metav1.Condition) {
	s.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&ScalewayCluster{}, &ScalewayClusterList{})
}
