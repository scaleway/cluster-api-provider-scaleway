package scope

import (
	"context"
	"fmt"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"golang.org/x/crypto/blake2b"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
	scwClient "github.com/scaleway/cluster-api-provider-scaleway/internal/service/scaleway/client"
)

const base36set = "0123456789abcdefghijklmnopqrstuvwxyz"

func nameWithSuffixes(name string, suffixes ...string) string {
	return strings.Join(append([]string{name}, suffixes...), "-")
}

func newScalewayClient(ctx context.Context, c client.Client, region, projectID string, secretRef client.ObjectKey) (*scwClient.Client, error) {
	r, err := scw.ParseRegion(region)
	if err != nil {
		return nil, fmt.Errorf("unable to parse region %q: %w", r, err)
	}

	secret := &corev1.Secret{}
	if err := c.Get(ctx, secretRef, secret); err != nil {
		return nil, fmt.Errorf("failed to get ScalewaySecret: %w", err)
	}

	sc, err := scwClient.New(r, projectID, secret.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to create Scaleway client from ScalewaySecret: %w", err)
	}

	return sc, nil
}

func newScalewayClientForScalewayCluster(ctx context.Context, c client.Client, sc *infrav1.ScalewayCluster) (*scwClient.Client, error) {
	return newScalewayClient(ctx, c, string(sc.Spec.Region), string(sc.Spec.ProjectID), types.NamespacedName{
		Namespace: sc.Namespace,
		Name:      sc.Spec.ScalewaySecretName,
	})
}

func newScalewayClientForScalewayManagedCluster(ctx context.Context, c client.Client, smc *infrav1.ScalewayManagedCluster) (*scwClient.Client, error) {
	return newScalewayClient(ctx, c, string(smc.Spec.Region), string(smc.Spec.ProjectID), types.NamespacedName{
		Namespace: smc.Namespace,
		Name:      smc.Spec.ScalewaySecretName,
	})
}

// base36TruncatedHash returns a consistent hash using blake2b
// and truncating the byte values to alphanumeric only
// of a fixed length specified by the consumer.
func base36TruncatedHash(str string, hashLen int) (string, error) {
	hasher, err := blake2b.New(hashLen, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create hash function: %w", err)
	}

	if _, err := hasher.Write([]byte(str)); err != nil {
		return "", fmt.Errorf("unable to write hash: %w", err)
	}

	return base36Truncate(hasher.Sum(nil)), nil
}

// base36Truncate returns a string that is base36 compliant
// It is not an encoding since it returns a same-length string
// for any byte value.
func base36Truncate(bytes []byte) string {
	var chars strings.Builder
	for _, bite := range bytes {
		idx := int(bite) % 36
		chars.WriteString(string(base36set[idx]))
	}

	return chars.String()
}
