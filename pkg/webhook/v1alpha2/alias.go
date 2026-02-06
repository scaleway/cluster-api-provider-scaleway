package webhook

import webhookv1 "github.com/scaleway/cluster-api-provider-scaleway/internal/webhook/v1alpha2"

// Following variables provide access to webhooks implemented in internal/webhook/v1alpha2.

var (
	SetupScalewayClusterWebhookWithManager             = webhookv1.SetupScalewayClusterWebhookWithManager
	SetupScalewayClusterTemplateWebhookWithManager     = webhookv1.SetupScalewayClusterTemplateWebhookWithManager
	SetupScalewayMachineWebhookWithManager             = webhookv1.SetupScalewayMachineWebhookWithManager
	SetupScalewayMachineTemplateWebhookWithManager     = webhookv1.SetupScalewayMachineTemplateWebhookWithManager
	SetupScalewayManagedClusterWebhookWithManager      = webhookv1.SetupScalewayManagedClusterWebhookWithManager
	SetupScalewayManagedControlPlaneWebhookWithManager = webhookv1.SetupScalewayManagedControlPlaneWebhookWithManager
	SetupScalewayManagedMachinePoolWebhookWithManager  = webhookv1.SetupScalewayManagedMachinePoolWebhookWithManager
)
