package v1alpha1

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

// PrivateNetworkParams allows to set the params of the Private Network.
type PrivateNetworkParams struct {
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

// CIDR is an IP address range in CIDR notation (for example, "10.0.0.0/8" or "fd00::/8").
// +kubebuilder:validation:XValidation:rule="isCIDR(self)",message="value must be a valid CIDR network address"
// +kubebuilder:validation:MaxLength:=43
// +kubebuilder:validation:MinLength:=1
type CIDR string
