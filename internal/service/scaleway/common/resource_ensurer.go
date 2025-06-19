package common

import (
	"context"
	"fmt"
	"slices"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

// ResourceReconciler defines a set of methods for managing resources in a desired state.
type ResourceReconciler[D, R any] interface {
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
	ShouldKeepResource(ctx context.Context, resource R, desired D) (bool, error)
}

// ResourceEnsurer is a utility that ensures a list of desired resources
type ResourceEnsurer[D, R any] struct {
	ResourceReconciler[D, R]
}

// Do ensures that the desired resources are provisioned. It also removes orphan resources.
func (e *ResourceEnsurer[D, R]) Do(ctx context.Context, desired []D) ([]R, error) {
	desiredResourcesByZone, err := e.indexDesiredResourcesByZone(desired)
	if err != nil {
		return nil, err
	}

	existingResources, err := e.ensureExistingResources(ctx, desiredResourcesByZone)
	if err != nil {
		return nil, err
	}

	createdResources, err := e.createMissingResources(ctx, existingResources, desiredResourcesByZone)
	if err != nil {
		return nil, err
	}

	return append(existingResources, createdResources...), nil
}

// indexDesiredResourcesByZone indexes desired resources by zone.
func (e *ResourceEnsurer[D, R]) indexDesiredResourcesByZone(desired []D) (map[scw.Zone][]D, error) {
	desiredResourcesByZone := make(map[scw.Zone][]D)

	for _, d := range desired {
		zone, err := e.GetDesiredZone(d)
		if err != nil {
			return nil, err
		}

		desiredResourcesByZone[zone] = append(desiredResourcesByZone[zone], d)
	}

	return desiredResourcesByZone, nil
}

// ensureExistingResources lists existing infra and removes everything that doesn't
// match currently desired resources.
func (e *ResourceEnsurer[D, R]) ensureExistingResources(
	ctx context.Context,
	desiredResourcesByZone map[scw.Zone][]D,
) ([]R, error) {
	resources, err := e.ListResources(ctx)
	if err != nil {
		return nil, err
	}

	keptResources := make([]R, 0)

	for _, resource := range resources {
		keep := false

		for i, desiredResource := range desiredResourcesByZone[e.GetResourceZone(resource)] {
			if e.GetResourceName(resource) != e.GetDesiredResourceName(i) {
				continue
			}

			// Writes to the keep variable outside the scope of this for-loop.
			keep, err = e.ShouldKeepResource(ctx, resource, desiredResource)
			if err != nil {
				return nil, err
			}

			// Continue looping to the next resource until we find one to keep.
			if !keep {
				continue
			}

			resource, err = e.UpdateResource(ctx, resource, desiredResource)
			if err != nil {
				return nil, fmt.Errorf("failed to update resource: %w", err)
			}

			break
		}

		if !keep {
			if err := e.DeleteResource(ctx, resource); err != nil {
				return nil, fmt.Errorf("failed to delete resource: %w", err)
			}

			continue
		}

		keptResources = append(keptResources, resource)
	}

	return keptResources, nil
}

func (e *ResourceEnsurer[D, R]) createMissingResources(
	ctx context.Context,
	existingResources []R,
	desiredResourcesByZone map[scw.Zone][]D,
) ([]R, error) {
	var resources []R

	for zone, desiredResources := range desiredResourcesByZone {
		for i, desiredResource := range desiredResources {
			desiredName := e.GetDesiredResourceName(i)

			// Skip if desired resource currently exists among existing resources.
			if slices.ContainsFunc(existingResources, func(resource R) bool {
				return desiredName == e.GetResourceName(resource) && e.GetResourceZone(resource) == zone
			}) {
				continue
			}

			resource, err := e.CreateResource(ctx, zone, desiredName, desiredResource)
			if err != nil {
				return nil, err
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}
