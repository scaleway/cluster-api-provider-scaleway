//go:generate ../../../../../bin/mockgen -destination client_mock.go -package mock_client -source ../interface.go -typed
//go:generate ../../../../../bin/mockgen -destination block_mock.go -package mock_client -source ../block.go -typed
//go:generate ../../../../../bin/mockgen -destination domain_mock.go -package mock_client -source ../domain.go -typed
//go:generate ../../../../../bin/mockgen -destination instance_mock.go -package mock_client -source ../instance.go -typed
//go:generate ../../../../../bin/mockgen -destination ipam_mock.go -package mock_client -source ../ipam.go -typed
//go:generate ../../../../../bin/mockgen -destination lb_mock.go -package mock_client -source ../lb.go -typed
//go:generate ../../../../../bin/mockgen -destination marketplace_mock.go -package mock_client -source ../marketplace.go -typed
//go:generate ../../../../../bin/mockgen -destination vpc_mock.go -package mock_client -source ../vpc.go -typed
//go:generate ../../../../../bin/mockgen -destination vpcgw_mock.go -package mock_client -source ../vpcgw.go -typed
//go:generate ../../../../../bin/mockgen -destination zones_mock.go -package mock_client -source ../zones.go -typed
package mock_client
