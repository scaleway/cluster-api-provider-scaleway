package vpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha1"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

type Scope interface {
	scope.Interface

	HasPrivateNetwork() bool
	IsVPCStatusSet() bool
	SetVPCStatus(privateNetworkID, VPCID string)
	PrivateNetworkParams() infrav1.PrivateNetworkParams
}

type Service struct {
	Scope
}

func New(s Scope) *Service {
	return &Service{s}
}

func (s Service) Name() string {
	return "vpc"
}

func (s *Service) Delete(ctx context.Context) error {
	if !s.HasPrivateNetwork() {
		return nil
	}

	params := s.PrivateNetworkParams()

	// User has provided his private network, we should not touch it.
	if params.ID != nil {
		return nil
	}

	pn, err := s.Cloud().FindPrivateNetwork(
		ctx,
		s.ResourceTags(),
		params.VPCID,
	)
	if err != nil {
		if errors.Is(err, client.ErrNoItemFound) {
			return nil
		}

		return fmt.Errorf("failed to find Private Network by name: %w", err)
	}

	if err := s.Cloud().CleanAvailableIPs(ctx, pn.ID); err != nil {
		return fmt.Errorf("failed to clean available IPs in IPAM: %w", err)
	}

	if err := s.Cloud().DeletePrivateNetwork(ctx, pn.ID); err != nil {
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

	// If the VPC and Private Network IDs are already set in the status, we don't need to do anything.
	if s.IsVPCStatusSet() {
		return nil
	}

	params := s.PrivateNetworkParams()

	var err error
	var pn *vpc.PrivateNetwork

	if pnID := params.ID; pnID != nil {
		pn, err = s.Cloud().GetPrivateNetwork(ctx, *pnID)
		if err != nil {
			return fmt.Errorf("failed to get existing Private Network: %w", err)
		}
	} else {
		pn, err = s.getOrCreatePN(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to get or create Private Network: %w", err)
		}
	}

	s.SetVPCStatus(pn.ID, pn.VpcID)

	return nil
}

func (s *Service) getOrCreatePN(ctx context.Context, params infrav1.PrivateNetworkParams) (*vpc.PrivateNetwork, error) {
	pn, err := s.Cloud().FindPrivateNetwork(
		ctx,
		s.ResourceTags(),
		params.VPCID,
	)
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return nil, err
	}

	if pn == nil {
		pn, err = s.Cloud().CreatePrivateNetwork(
			ctx,
			s.ResourceName(),
			params.VPCID,
			params.Subnet,
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
