package vpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

type Service struct {
	*scope.Cluster
}

func New(clusterScope *scope.Cluster) *Service {
	return &Service{Cluster: clusterScope}
}

func (s Service) Name() string {
	return "vpc"
}

func (s *Service) Delete(ctx context.Context) error {
	if !s.ShouldManagePrivateNetwork() {
		return nil
	}

	pn, err := s.ScalewayClient.FindPrivateNetwork(
		ctx,
		s.ResourceTags(),
		s.ScalewayCluster.Spec.Network.PrivateNetwork.VPCID,
	)
	if err != nil {
		if errors.Is(err, client.ErrNoItemFound) {
			return nil
		}

		return fmt.Errorf("failed to find Private Network by name: %w", err)
	}

	if err := s.ScalewayClient.CleanAvailableIPs(ctx, pn.ID); err != nil {
		return fmt.Errorf("failed to clean available IPs in IPAM: %w", err)
	}

	if err := s.ScalewayClient.DeletePrivateNetwork(ctx, pn.ID); err != nil {
		// Sometimes, we still need to wait a little for all ressources to be removed
		// from the Private Network. As a result, we need to handle this error:
		// scaleway-sdk-go: precondition failed: resource is still in use, Private Network must be empty to be deleted
		if client.IsPreconditionFailedError(err) {
			return scaleway.WithTransientError(err, 5*time.Second)
		}

		return fmt.Errorf("failed to delete Private Network: %w", err)
	}

	return nil
}

func (s *Service) Reconcile(ctx context.Context) error {
	if !s.HasPrivateNetwork() {
		return nil
	}

	if s.ScalewayCluster.Status.Network != nil &&
		s.ScalewayCluster.Status.Network.PrivateNetworkID != nil &&
		s.ScalewayCluster.Status.Network.VPCID != nil {
		// If the VPC and Private Network IDs are already set in the status, we don't need to do anything.
		return nil
	}

	var err error
	var pn *vpc.PrivateNetwork

	if s.ShouldManagePrivateNetwork() {
		pn, err = s.getOrCreatePN(ctx)
		if err != nil {
			return fmt.Errorf("failed to get or create Private Network: %w", err)
		}
	} else {
		pn, err = s.ScalewayClient.GetPrivateNetwork(ctx, *s.ScalewayCluster.Spec.Network.PrivateNetwork.ID)
		if err != nil {
			return fmt.Errorf("failed to get existing Private Network: %w", err)
		}
	}

	s.SetStatusPrivateNetworkID(pn.ID)
	s.SetStatusVPCID(pn.VpcID)

	return nil
}

func (s *Service) getOrCreatePN(ctx context.Context) (*vpc.PrivateNetwork, error) {
	pn, err := s.ScalewayClient.FindPrivateNetwork(
		ctx,
		s.ResourceTags(),
		s.ScalewayCluster.Spec.Network.PrivateNetwork.VPCID,
	)
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return nil, err
	}

	if pn == nil {
		pn, err = s.ScalewayClient.CreatePrivateNetwork(
			ctx,
			s.ResourceName(),
			s.ScalewayCluster.Spec.Network.PrivateNetwork.VPCID,
			s.ScalewayCluster.Spec.Network.PrivateNetwork.Subnet,
			s.ResourceTags(),
		)
		if err != nil {
			return nil, err
		}
	}

	if !pn.DHCPEnabled {
		return nil, scaleway.WithTerminalError(
			fmt.Errorf("Private Network with ID %s is not supported: DHCP is not enabled", pn.ID),
		)
	}

	return pn, nil
}
