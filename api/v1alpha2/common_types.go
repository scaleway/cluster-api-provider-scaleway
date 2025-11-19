package v1alpha2

// UUID is a valid UUID for a Scaleway resource.
// +kubebuilder:validation:Pattern="^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
// +kubebuilder:validation:MinLength=36
// +kubebuilder:validation:MaxLength=36
type UUID string

// CIDR is an IP address range in CIDR notation (for example, "10.0.0.0/8" or "fd00::/8").
// +kubebuilder:validation:XValidation:rule="isCIDR(self)",message="value must be a valid CIDR network address"
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=43
type CIDR string

// IPv4 is a valid IPv4.
// +kubebuilder:validation:Format=ipv4
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=15
type IPv4 string

// ScalewayRegion is a Scaleway region (e.g. fr-par).
// +kubebuilder:validation:Pattern="^[a-z]{2}-[a-z]{3}$"
// +kubebuilder:validation:MinLength=6
// +kubebuilder:validation:MaxLength=6
type ScalewayRegion string

// ScalewayZone is a Scaleway zone (e.g. fr-par-1).
// +kubebuilder:validation:Pattern="^[a-z]{2}-[a-z]{3}-[0-9]{0,2}$"
// +kubebuilder:validation:MinLength=8
// +kubebuilder:validation:MaxLength=9
type ScalewayZone string

// PrivateNetwork allows to reference an existing Private Network or configure
// the parameters of the auto-created Private Network.
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:XValidation:rule="has(self.vpcID) == has(oldSelf.vpcID)",message="vpcID cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.id) == has(oldSelf.id)",message="id cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="has(self.subnet) == has(oldSelf.subnet)",message="subnet cannot be added or removed"
// +kubebuilder:validation:XValidation:rule="!has(self.id) || has(self.id) != has(self.vpcID)",message="id and vpcID cannot be set at the same time"
// +kubebuilder:validation:XValidation:rule="!has(self.id) || has(self.id) != has(self.subnet)",message="id and subnet cannot be set at the same time"
type PrivateNetwork struct {
	// id allows to reuse an existing Private Network instead of creating a new one.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ID UUID `json:"id,omitempty"`

	// vpcID defines the ID of the VPC where the new Private Network will be created if none is provided.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	VPCID UUID `json:"vpcID,omitempty"`

	// subnet defines a subnet for the Private Network. Only used on newly created Private Networks.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Subnet CIDR `json:"subnet,omitempty"`
}

// PublicGateway defines settings of the Public Gateway that will be created.
// +kubebuilder:validation:MinProperties=1
type PublicGateway struct {
	// type is a Public Gateway commercial offer type.
	// +optional
	// +kubebuilder:default="VPC-GW-S"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=20
	Type string `json:"type,omitempty"`

	// ip to use when creating a Public Gateway.
	// +optional
	IP IPv4 `json:"ip,omitempty"`

	// zone where to create the Public Gateway. Must be in the same region as the
	// cluster. Defaults to the first zone of the region.
	// +optional
	Zone ScalewayZone `json:"zone,omitempty"`
}
