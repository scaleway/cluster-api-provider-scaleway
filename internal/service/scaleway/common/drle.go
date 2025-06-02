package common

import (
	"context"
	"fmt"
	"slices"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

// DesiredResourceListManager is an interface for managing a list of desired resources.
type DesiredResourceListManager[D, R any] interface {
	// ListResources lists the resources that currently exist.
	ListResources(ctx context.Context) ([]R, error)
	// DeleteResource deletes a resource.
	DeleteResource(ctx context.Context, resource R) error
	// UpdateResource updates an existing resource (if needed).
	UpdateResource(ctx context.Context, resource R, desired D) (R, error)
	// CreateResource creates a new resource with the specified zone, name and specs.
	CreateResource(ctx context.Context, zone scw.Zone, name string, desired D) (R, error)

	// GetResourceZone returns the zone of a resource.
	GetResourceZone(resource R) scw.Zone
	// GetResourceName returns the name of a resource.
	GetResourceName(resource R) string

	// GetDesiredZone returns the desired zone for a resource according to its specs.
	GetDesiredZone(desired D) (scw.Zone, error)
	// GetDesiredResourceName returns the desired name for a resource at index i.
	GetDesiredResourceName(i int) string

	// ShouldKeepResource returns true if a resource should be kept because it
	// matches the desired specs.
	ShouldKeepResource(resource R, desired D) bool
}

// DesiredResourceListEnsure contains a DesiredResourceListManager.
type DesiredResourceListEnsure[D, R any] struct {
	DesiredResourceListManager[D, R]
}

// Do ensures that the desired resources are provisioned. It also removes orphan resources.
func (drle *DesiredResourceListEnsure[D, R]) Do(ctx context.Context, desired []D) ([]R, error) {
	desiredResourcesByZone, err := drle.indexDesiredResourcesByZone(desired)
	if err != nil {
		return nil, err
	}

	existingResources, err := drle.ensureExistingResources(ctx, desiredResourcesByZone)
	if err != nil {
		return nil, err
	}

	createdResources, err := drle.createMissingResources(ctx, existingResources, desiredResourcesByZone)
	if err != nil {
		return nil, err
	}

	return append(existingResources, createdResources...), nil
}

// indexDesiredResourcesByZone indexes desired resources by zone.
func (drle *DesiredResourceListEnsure[D, R]) indexDesiredResourcesByZone(desired []D) (map[scw.Zone][]D, error) {
	desiredResourcesByZone := make(map[scw.Zone][]D)

	for _, d := range desired {
		zone, err := drle.GetDesiredZone(d)
		if err != nil {
			return nil, err
		}

		desiredResourcesByZone[zone] = append(desiredResourcesByZone[zone], d)
	}

	return desiredResourcesByZone, nil
}

// ensureExistingResources lists existing infra and removes everything that doesn't
// match currently desired resources.
func (drle *DesiredResourceListEnsure[D, R]) ensureExistingResources(
	ctx context.Context,
	desiredResourcesByZone map[scw.Zone][]D,
) ([]R, error) {
	resources, err := drle.ListResources(ctx)
	if err != nil {
		return nil, err
	}

	keptResources := make([]R, 0)

	for _, resource := range resources {
		keep := false

		for i, desiredResource := range desiredResourcesByZone[drle.GetResourceZone(resource)] {
			if drle.GetResourceName(resource) != drle.GetDesiredResourceName(i) {
				continue
			}

			if !drle.ShouldKeepResource(resource, desiredResource) {
				continue
			}

			keep = true

			resource, err = drle.UpdateResource(ctx, resource, desiredResource)
			if err != nil {
				return nil, fmt.Errorf("failed to update resource: %w", err)
			}

			break
		}

		if !keep {
			if err := drle.DeleteResource(ctx, resource); err != nil {
				return nil, fmt.Errorf("failed to delete resource: %w", err)
			}

			continue
		}

		keptResources = append(keptResources, resource)
	}

	return keptResources, nil
}

func (drle *DesiredResourceListEnsure[D, R]) createMissingResources(
	ctx context.Context,
	existingResources []R,
	desiredResourcesByZone map[scw.Zone][]D,
) ([]R, error) {
	var resources []R

	for zone, desiredResources := range desiredResourcesByZone {
		for i, desiredResource := range desiredResources {
			desiredName := drle.GetDesiredResourceName(i)

			// Skip if desired resource currently exists among existing resources.
			if slices.ContainsFunc(existingResources, func(resource R) bool {
				return desiredName == drle.GetResourceName(resource) && drle.GetResourceZone(resource) == zone
			}) {
				continue
			}

			resource, err := drle.CreateResource(ctx, zone, desiredName, desiredResource)
			if err != nil {
				return nil, err
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}
