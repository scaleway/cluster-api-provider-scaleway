package cluster

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/cluster-api/util/secret"

	infrav1 "github.com/scaleway/cluster-api-provider-scaleway/api/v1alpha2"
)

func (s *Service) reconcileKubeconfig(ctx context.Context, cluster *k8s.Cluster, getKubeconfig kubeconfigGetter) error {
	clusterRef := types.NamespacedName{
		Name:      s.Cluster.Name,
		Namespace: s.Cluster.Namespace,
	}

	configSecret, err := secret.GetFromNamespacedName(ctx, s.Client, clusterRef, secret.Kubeconfig)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("getting kubeconfig secret %s: %w", clusterRef, err)
		}

		if createErr := s.createCAPIKubeconfigSecret(ctx, cluster, getKubeconfig, &clusterRef); createErr != nil {
			return fmt.Errorf("creating kubeconfig secret: %w", createErr)
		}
	} else if updateErr := s.updateCAPIKubeconfigSecret(ctx, configSecret); updateErr != nil {
		return fmt.Errorf("updating kubeconfig secret: %w", err)
	}

	return nil
}

func (s *Service) reconcileAdditionalKubeconfigs(ctx context.Context, cluster *k8s.Cluster, getKubeconfig kubeconfigGetter) error {
	clusterRef := types.NamespacedName{
		Name:      s.Cluster.Name + "-user",
		Namespace: s.Cluster.Namespace,
	}

	// Create the additional kubeconfig for users. This doesn't need updating on every sync
	if _, err := secret.GetFromNamespacedName(ctx, s.Client, clusterRef, secret.Kubeconfig); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("getting kubeconfig (user) secret %s: %w", clusterRef, err)
		}

		createErr := s.createUserKubeconfigSecret(ctx, cluster, getKubeconfig, &clusterRef, s.Cluster.Name)
		if createErr != nil {
			return fmt.Errorf("creating additional kubeconfig secret: %w", createErr)
		}
	}

	return nil
}

func (s *Service) createUserKubeconfigSecret(
	ctx context.Context,
	cluster *k8s.Cluster,
	getKubeconfig kubeconfigGetter,
	clusterRef *types.NamespacedName,
	clusterName string,
) error {
	controllerOwnerRef := *metav1.NewControllerRef(s.ScalewayManagedControlPlane, infrav1.GroupVersion.WithKind("ScalewayManagedControlPlane"))

	contextName := s.getKubeConfigContextName(false)

	kc, err := getKubeconfig()
	if err != nil {
		return err
	}

	cfg, err := s.createBaseKubeConfig(contextName, cluster, kc)
	if err != nil {
		return fmt.Errorf("creating base kubeconfig: %w", err)
	}

	execConfig := &api.ExecConfig{
		APIVersion:      "client.authentication.k8s.io/v1",
		Command:         "scw",
		Args:            []string{"k8s", "exec-credential"},
		InteractiveMode: api.NeverExecInteractiveMode,
		InstallHint:     "Install scaleway CLI for use with kubectl by following\n		https://cli.scaleway.com/#installation",
	}
	cfg.AuthInfos = map[string]*api.AuthInfo{
		contextName: {
			Exec: execConfig,
		},
	}

	out, err := clientcmd.Write(*cfg)
	if err != nil {
		return fmt.Errorf("serialize kubeconfig to yaml: %w", err)
	}

	kubeconfigSecret := kubeconfig.GenerateSecretWithOwner(*clusterRef, out, controllerOwnerRef)
	kubeconfigSecret.Labels[clusterv1.ClusterNameLabel] = clusterName

	if err := s.Client.Create(ctx, kubeconfigSecret); err != nil {
		return fmt.Errorf("creating secret: %w", err)
	}

	return nil
}

func (s *Service) createCAPIKubeconfigSecret(ctx context.Context, cluster *k8s.Cluster, getKubeconfig kubeconfigGetter, clusterRef *types.NamespacedName) error {
	controllerOwnerRef := *metav1.NewControllerRef(s.ScalewayManagedControlPlane, infrav1.GroupVersion.WithKind("ScalewayManagedControlPlane"))

	contextName := s.getKubeConfigContextName(false)

	kc, err := getKubeconfig()
	if err != nil {
		return err
	}

	cfg, err := s.createBaseKubeConfig(contextName, cluster, kc)
	if err != nil {
		return fmt.Errorf("creating base kubeconfig: %w", err)
	}

	cfg.AuthInfos = map[string]*api.AuthInfo{
		contextName: {
			Token: s.ScalewayClient.GetSecretKey(),
		},
	}

	out, err := clientcmd.Write(*cfg)
	if err != nil {
		return fmt.Errorf("serialize kubeconfig to yaml: %w", err)
	}

	kubeconfigSecret := kubeconfig.GenerateSecretWithOwner(*clusterRef, out, controllerOwnerRef)
	if err := s.Client.Create(ctx, kubeconfigSecret); err != nil {
		return fmt.Errorf("creating secret: %w", err)
	}

	return nil
}

func (s *Service) updateCAPIKubeconfigSecret(ctx context.Context, configSecret *corev1.Secret) error {
	data, ok := configSecret.Data[secret.KubeconfigDataName]
	if !ok {
		return fmt.Errorf("missing key %q in secret data", secret.KubeconfigDataName)
	}

	config, err := clientcmd.Load(data)
	if err != nil {
		return fmt.Errorf("failed to convert kubeconfig Secret into a clientcmdapi.Config: %w", err)
	}

	contextName := s.getKubeConfigContextName(false)

	if config.AuthInfos[contextName] == nil {
		return nil
	}

	if config.AuthInfos[contextName].Token == s.ScalewayClient.GetSecretKey() {
		return nil
	}

	config.AuthInfos[contextName].Token = s.ScalewayClient.GetSecretKey()

	out, err := clientcmd.Write(*config)
	if err != nil {
		return fmt.Errorf("failed to serialize config to yaml: %w", err)
	}

	configSecret.Data[secret.KubeconfigDataName] = out

	if err := s.Client.Update(ctx, configSecret); err != nil {
		return fmt.Errorf("updating kubeconfig secret: %w", err)
	}

	return nil
}

func (s *Service) createBaseKubeConfig(contextName string, cluster *k8s.Cluster, kc *k8s.Kubeconfig) (*api.Config, error) {
	b64CACert, err := kc.GetCertificateAuthorityData()
	if err != nil {
		return nil, err
	}
	certData, err := base64.StdEncoding.DecodeString(b64CACert)
	if err != nil {
		return nil, fmt.Errorf("decoding cluster CA cert: %w", err)
	}

	cfg := &api.Config{
		APIVersion: api.SchemeGroupVersion.Version,
		Clusters: map[string]*api.Cluster{
			contextName: {
				Server:                   s.ClusterEndpoint(cluster),
				CertificateAuthorityData: certData,
			},
		},
		Contexts: map[string]*api.Context{
			contextName: {
				Cluster:  contextName,
				AuthInfo: contextName,
			},
		},
		CurrentContext: contextName,
	}

	return cfg, nil
}

func (s *Service) getKubeConfigContextName(isUser bool) string {
	contextName := fmt.Sprintf("scw_%s_%s_%s", s.ScalewayManagedCluster.Spec.ProjectID, s.ScalewayManagedCluster.Spec.Region, s.ClusterName())
	if isUser {
		contextName += "-user"
	}
	return contextName
}
