package vpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cluster-api/util/conditions"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/scope"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway"
	"github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
)

type Scope interface {
	scope.Interface
	conditions.Setter

	HasPrivateNetwork() bool
	IsVPCStatusSet() bool
	SetVPCStatus(privateNetworkID, VPCID string)
	PrivateNetwork() infrav1.PrivateNetwork
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

	params := s.PrivateNetwork()

	// User has provided his private network, we should not touch it.
	if params.ID != "" {
		return nil
	}

	var vpcID *string
	if params.VPCID != "" {
		vpcID = ptr.To(string(params.VPCID))
	}

	pn, err := s.Cloud().FindPrivateNetwork(ctx, s.ResourceTags(), vpcID)
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
		conditions.Set(s, metav1.Condition{
			Type:   infrav1.PrivateNetworkReadyCondition,
			Status: metav1.ConditionTrue,
			Reason: infrav1.NoPrivateNetworkReason,
		})
		return nil
	}

	// If status is already configured, we don't need to do anything.
	if conditions.IsTrue(s, infrav1.PrivateNetworkReadyCondition) && s.IsVPCStatusSet() {
		return nil
	}

	params := s.PrivateNetwork()

	var err error
	var pn *vpc.PrivateNetwork

	if pnID := params.ID; pnID != "" {
		pn, err = s.Cloud().GetPrivateNetwork(ctx, string(pnID))
		if err != nil {
			conditions.Set(s, metav1.Condition{
				Type:    infrav1.PrivateNetworkReadyCondition,
				Status:  metav1.ConditionFalse,
				Reason:  infrav1.PrivateNetworkNotFoundReason,
				Message: err.Error(),
			})
			return fmt.Errorf("failed to get existing Private Network: %w", err)
		}
	} else {
		pn, err = s.getOrCreatePN(ctx, params)
		if err != nil {
			conditions.Set(s, metav1.Condition{
				Type:    infrav1.PrivateNetworkReadyCondition,
				Status:  metav1.ConditionFalse,
				Reason:  infrav1.CreationFailedReason,
				Message: err.Error(),
			})
			return fmt.Errorf("failed to get or create Private Network: %w", err)
		}
	}

	s.SetVPCStatus(pn.ID, pn.VpcID)

	conditions.Set(s, metav1.Condition{
		Type:   infrav1.PrivateNetworkReadyCondition,
		Status: metav1.ConditionTrue,
		Reason: infrav1.ReadyReason,
	})

	return nil
}

func (s *Service) getOrCreatePN(ctx context.Context, params infrav1.PrivateNetwork) (*vpc.PrivateNetwork, error) {
	var vpcID *string
	if params.VPCID != "" {
		vpcID = ptr.To(string(params.VPCID))
	}

	pn, err := s.Cloud().FindPrivateNetwork(ctx, s.ResourceTags(), vpcID)
	if err := utilerrors.FilterOut(err, client.IsNotFoundError); err != nil {
		return nil, err
	}

	var subnet *string
	if params.Subnet != "" {
		subnet = ptr.To(string(params.Subnet))
	}

	if pn == nil {
		pn, err = s.Cloud().CreatePrivateNetwork(ctx, s.ResourceName(), vpcID, subnet, s.ResourceTags())
		if err != nil {
			return nil, err
		}
	}

	if !pn.DHCPEnabled {
		return nil, fmt.Errorf("Private Network with ID %s is not supported: DHCP is not enabled", pn.ID)
	}

	return pn, nil
}
