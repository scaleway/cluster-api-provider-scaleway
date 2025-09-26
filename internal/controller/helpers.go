package controller

import (
	"context"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

var (
	scalewaySecretOwnerAPIVersion = infrav1.GroupVersion.String()
	scalewaySecretOwnerKinds      = []string{"ScalewayCluster", "ScalewayManagedCluster"}
)

// claimScalewaySecret adds an object as owner of a secret. It also adds a finalizer
// (if not present already) to prevent the removal of the secret.
func claimScalewaySecret(ctx context.Context, c client.Client, owner client.Object, secretName string) error {
	gvk, err := apiutil.GVKForObject(owner, c.Scheme())
	if err != nil {
		return fmt.Errorf("failed to get GVK for owner: %w", err)
	}

	if !slices.Contains(scalewaySecretOwnerKinds, gvk.Kind) {
		return fmt.Errorf("object with kind %s cannot own scaleway secret", gvk.Kind)
	}

	secret := &corev1.Secret{}
	if err := c.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: owner.GetNamespace(),
	}, secret); err != nil {
		return err
	}

	secretHelper, err := patch.NewHelper(secret, c)
	if err != nil {
		return fmt.Errorf("failed to create patch helper for secret: %w", err)
	}

	controllerutil.AddFinalizer(secret, SecretFinalizer)

	if err := controllerutil.SetOwnerReference(owner, secret, c.Scheme()); err != nil {
		return fmt.Errorf("failed to set owner reference for secret %s: %w", secret.Name, err)
	}

	return secretHelper.Patch(ctx, secret)
}

// releaseScalewaySecret removes an object as owner of a secret. It also removes
// the finalizer it there is no owner anymore.
func releaseScalewaySecret(ctx context.Context, c client.Client, owner client.Object, secretName string) error {
	gvk, err := apiutil.GVKForObject(owner, c.Scheme())
	if err != nil {
		return fmt.Errorf("failed to get GVK for owner: %w", err)
	}

	if !slices.Contains(scalewaySecretOwnerKinds, gvk.Kind) {
		return fmt.Errorf("object with kind %s cannot own scaleway secret", gvk.Kind)
	}

	secret := &corev1.Secret{}
	if err := c.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: owner.GetNamespace(),
	}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	secretHelper, err := patch.NewHelper(secret, c)
	if err != nil {
		return fmt.Errorf("failed to create patch helper for secret: %w", err)
	}

	hasOwnerReference, err := controllerutil.HasOwnerReference(secret.OwnerReferences, owner, c.Scheme())
	if err != nil {
		return fmt.Errorf("failed to check owner refenrece for secret %s: %w", secret.Name, err)
	}

	if hasOwnerReference {
		if err := controllerutil.RemoveOwnerReference(owner, secret, c.Scheme()); err != nil {
			return fmt.Errorf("failed to remove owner reference for secret %s: %w", secret.Name, err)
		}
	}

	if !util.HasOwner(secret.OwnerReferences, scalewaySecretOwnerAPIVersion, scalewaySecretOwnerKinds) {
		controllerutil.RemoveFinalizer(secret, SecretFinalizer)
	}

	return secretHelper.Patch(ctx, secret)
}

func migrateFinalizer(o client.Object, old, finalizer string) bool {
	// Attempt to remove old finalizer.
	if !controllerutil.RemoveFinalizer(o, old) {
		return false
	}

	// Add the up-to-date finalizer.
	controllerutil.AddFinalizer(o, finalizer)
	return true
}
