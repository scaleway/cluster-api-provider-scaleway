package client

// Interface of the scaleway-sdk-go wrapper to access Scaleway Product APIs in
// a specific region and project.
type Interface interface {
	Block
	Domain
	Instance
	IPAM
	LB
	Marketplace
	VPC
	VPCGW
	Zones
}
