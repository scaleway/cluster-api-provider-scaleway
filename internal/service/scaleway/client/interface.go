package client

// Interface of the scaleway-sdk-go wrapper to access Scaleway Product APIs in
// a specific region and project.
type Interface interface {
	Block
	Config
	Domain
	Instance
	IPAM
	K8s
	LB
	Marketplace
	VPC
	VPCGW
	Zones
}
