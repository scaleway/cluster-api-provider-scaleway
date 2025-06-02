package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const ClusterFinalizer = "scalewaycluster.infrastructure.cluster.x-k8s.io/sc-protection"

// ScalewayClusterSpec defines the desired state of ScalewayCluster.
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.controlPlaneEndpoint) || has(self.controlPlaneEndpoint)", message="controlPlaneEndpoint is required once set"
type ScalewayClusterSpec struct {
	// Region represents the region where the cluster will be hosted.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=10
	Region string `json:"region"`

	// Network contains network related options for the cluster.
	// +optional
	Network *NetworkSpec `json:"network,omitempty"`

	// ScalewaySecretName is the name of the secret that contains the Scaleway client parameters.
	// The following keys are required: SCW_ACCESS_KEY, SCW_SECRET_KEY, SCW_DEFAULT_PROJECT_ID.
	// The following key is optional: SCW_API_URL.
	ScalewaySecretName string `json:"scalewaySecretName"`

	// FailureDomains is a list of failure domains where the control-plane nodes will be created.
	// Failure domains correspond to Scaleway zones inside the cluster region (e.g. fr-par-1).
	// +listType=set
	// +optional
	FailureDomains []string `json:"failureDomains,omitempty"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`
}

// NetworkSpec defines network specific settings.
// +kubebuilder:validation:XValidation:rule="!has(self.controlPlaneExtraLoadBalancers) || has(self.controlPlaneDNS)",message="controlPlaneDNS is required when controlPlaneExtraLoadBalancers is set"
// +kubebuilder:validation:XValidation:rule="has(self.controlPlaneDNS) == has(oldSelf.controlPlaneDNS)",message="controlPlaneDNS cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.privateNetwork) == has(oldSelf.privateNetwork)",message="privateNetwork cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="!has(self.publicGateways) || has(self.privateNetwork) && self.privateNetwork.enabled",message="privateNetwork is required when publicGateways is set"
type NetworkSpec struct {
	// ControlPlaneLoadBalancer contains loadbalancer settings.
	// +optional
	ControlPlaneLoadBalancer *ControlPlaneLoadBalancerSpec `json:"controlPlaneLoadBalancer,omitempty"`

	// ControlPlaneExtraLoadBalancers allows configuring additional LoadBalancers.
	// Because Scaleway LoadBalancers are currently zonal resources, you may set
	// up to 3 additional LoadBalancers for achieving regional redundancy. It is
	// mandatory to set the controlPlaneDNS field when you do so.
	// This may be removed in the future, when Scaleway supports regional LoadBalancers.
	// +kubebuilder:validation:MaxItems=3
	// +optional
	ControlPlaneExtraLoadBalancers []LoadBalancerSpec `json:"controlPlaneExtraLoadBalancers,omitempty"`

	// ControlPlaneDNS allows configuring a Scaleway Domain DNS Zone.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	ControlPlaneDNS *ControlPlaneDNSSpec `json:"controlPlaneDNS,omitempty"`

	// PrivateNetwork allows attaching machines of the cluster to a Private Network.
	// +optional
	PrivateNetwork *PrivateNetworkSpec `json:"privateNetwork,omitempty"`

	// PublicGateways allows to create Public Gateways that will be attached to the
	// Private Network of the cluster.
	// +kubebuilder:validation:MaxItems=6
	// +optional
	PublicGateways []PublicGatewaySpec `json:"publicGateways,omitempty"`
}

// LoadBalancerSpec defines loadbalancer parameters.
type LoadBalancerSpec struct {
	// Zone where to create the loadbalancer. Must be in the same region as the
	// cluster. Defaults to the first zone of the region.
	// +optional
	Zone *string `json:"zone,omitempty"`

	// Load Balancer commercial offer type.
	// +kubebuilder:default="LB-S"
	// +optional
	Type *string `json:"type,omitempty"`

	// IP to use when creating a loadbalancer.
	// +kubebuilder:validation:Format=ipv4
	// +optional
	IP *string `json:"ip,omitempty"`
}

// ControlPlaneLoadBalancerSpec defines control-plane loadbalancer settings for the cluster.
// +kubebuilder:validation:XValidation:rule="has(self.port) == has(oldSelf.port)",message="port cannot be added or removed"
type ControlPlaneLoadBalancerSpec struct {
	// +kubebuilder:validation:XValidation:rule="has(self.ip) == has(oldSelf.ip)",message="ip cannot be added or removed"
	// +kubebuilder:validation:XValidation:rule="!has(oldSelf.ip) || self.ip == oldSelf.ip",message="ip is immutable"
	// +kubebuilder:validation:XValidation:rule="has(self.zone) == has(oldSelf.zone)",message="zone cannot be added or removed"
	// +kubebuilder:validation:XValidation:rule="!has(oldSelf.zone) || self.zone == oldSelf.zone",message="zone is immutable"
	LoadBalancerSpec `json:",inline"`

	// Port configured on the Load Balancer. It must be valid port range (1-65535).
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *int32 `json:"port,omitempty"`

	// AllowedRanges allows to set a list of allowed IP ranges that can access
	// the cluster through the loadbalancer. When unset, all IP ranges are allowed.
	// To allow the cluster to work properly, public IPs of nodes and Public
	// Gateways will automatically be allowed. However, if this field is set,
	// you MUST manually allow IPs of the nodes of your management cluster.
	// +kubebuilder:validation:MaxItems=30
	// +listType=set
	// +optional
	AllowedRanges []CIDR `json:"allowedRanges,omitempty"`
}

// CIDR is an IP address range in CIDR notation (for example, "10.0.0.0/8" or "fd00::/8").
// +kubebuilder:validation:XValidation:rule="isCIDR(self)",message="value must be a valid CIDR network address"
// +kubebuilder:validation:MaxLength:=43
// +kubebuilder:validation:MinLength:=1
type CIDR string

type ControlPlaneDNSSpec struct {
	// Domain is the DNS Zone that this record should live in. It must be pre-existing in your Scaleway account.
	// The format must be a string that conforms to the definition of a subdomain in DNS (RFC 1123).
	// +kubebuilder:validation:Pattern:=^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$
	Domain string `json:"domain"`
	// Name is the DNS short name of the record (non-FQDN). The format must consist of
	// alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.
	// +kubebuilder:validation:Pattern:=^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$
	Name string `json:"name"`
}

// PrivateNetworkSpec defines Private Network settings for the cluster.
// +kubebuilder:validation:XValidation:rule="has(self.vpcID) == has(oldSelf.vpcID)",message="vpcID cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.id) == has(oldSelf.id)",message="id cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.subnet) == has(oldSelf.subnet)",message="subnet cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.id) && !has(self.subnet) || !has(self.id)",message="subnet cannot be set when id is set"
// +kubebuilder:validation:XValidation:rule="has(self.id) && !has(self.vpcID) || !has(self.id)",message="vpcID cannot be set when id is set"
type PrivateNetworkSpec struct {
	// Set to true to automatically attach machines to a Private Network.
	// The Private Network is automatically created if no existing Private
	// Network ID is provided.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Enabled bool `json:"enabled"`

	// Set a Private Network ID to reuse an existing Private Network.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	ID *string `json:"id,omitempty"`

	// Set the VPC ID where the new Private Network will be created.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	VPCID *string `json:"vpcID,omitempty"`

	// Optional subnet for the Private Network. Only used on newly created Private Networks.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	Subnet *string `json:"subnet,omitempty"`
}

// PublicGatewaySpec defines Public Gateway settings for the cluster.
type PublicGatewaySpec struct {
	// Public Gateway commercial offer type.
	// +kubebuilder:default="VPC-GW-S"
	// +optional
	Type *string `json:"type,omitempty"`

	// IP to use when creating a Public Gateway.
	// +kubebuilder:validation:Format=ipv4
	// +optional
	IP *string `json:"ip,omitempty"`

	// Zone where to create the Public Gateway. Must be in the same region as the
	// cluster. Defaults to the first zone of the region.
	// +optional
	Zone *string `json:"zone,omitempty"`
}

// ScalewayClusterStatus defines the observed state of ScalewayCluster.
type ScalewayClusterStatus struct {
	// Ready denotes that the Scaleway cluster infrastructure is fully provisioned.
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed.
	// +optional
	Ready bool `json:"ready"`

	// Network contains information about network resources of the cluster.
	// +optional
	Network *NetworkStatus `json:"network,omitempty"`

	// FailureDomains is a list of failure domain objects synced from the infrastructure provider.
	// +optional
	FailureDomains clusterv1.FailureDomains `json:"failureDomains,omitempty"`
}

// NetworkStatus contains information about network resources of the cluster.
type NetworkStatus struct {
	// PrivateNetworkID is set if the cluster has an associated Private Network.
	// +optional
	PrivateNetworkID *string `json:"privateNetworkID,omitempty"`

	// PublicGatewayIDs is a list of Public Gateway IDs.
	// +optional
	PublicGatewayIDs []string `json:"publicGatewayIDs,omitempty"`

	// LoadBalancerIP is the public IP of the cluster control-plane.
	// +optional
	LoadBalancerIP *string `json:"loadBalancerIP,omitempty"`

	// ExtraLoadBalancerIPs is a list of IPs of the extra loadbalancers.
	// +optional
	ExtraLoadBalancerIPs []string `json:"extraLoadBalancerIPs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Host",type="string",JSONPath=".spec.controlPlaneEndpoint.host",description="Host of the control plane"
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".spec.controlPlaneEndpoint.port",description="Port of the control plane"
// +kubebuilder:printcolumn:name="Region",type="string",JSONPath=".spec.region",description="Region of the cluster"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Ready is true when the cluster is fully provisioned"
// +kubebuilder:resource:path=scalewayclusters,scope=Namespaced,categories=cluster-api,shortName=sc
// +kubebuilder:storageversion

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
