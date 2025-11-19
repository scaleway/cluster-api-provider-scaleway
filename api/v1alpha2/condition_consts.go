package v1alpha2

const (
	// PrivateNetworkReadyCondition surfaces details about the current status of the private network.
	PrivateNetworkReadyCondition = "PrivateNetworkReady"

	// PublicGatewaysReadyCondition surfaces details about the current status of the public gateways.
	PublicGatewaysReadyCondition = "PublicGatewaysReady"
)

const (
	// ReadyReason surfaces when the resource is ready.
	ReadyReason = "Ready"

	// ReadyReason surfaces when the resource is not ready.
	NotReadyReason = "NotReady"

	// ProvisionedReason surfaces when the resource is provisioned.
	ProvisionedReason = "Provisioned"

	// NoPrivateNetworkReason surfaces when no private network is defined in the spec.
	NoPrivateNetworkReason = "NoPrivateNetwork"

	// PrivateNetworkNotFoundReason surfaces when the provided private network cannot be found.
	PrivateNetworkNotFoundReason = "PrivateNetworkNotfound"

	// CreationFailedReason surfaces when the resource creation failed.
	CreationFailedReason = "CreationFailed"

	// ReconciliationFailedReason surfaces when the resource reconciliation failed.
	ReconciliationFailedReason = "ReconciliationFailed"

	// PrivateNetworkAttachmentFailedReason surfaces when the attachment of resources to the private network failed.
	PrivateNetworkAttachmentFailedReason = "PrivateNetworkAttachmentFailed"

	// InternalErrorReason surfaces unexpected errors reporting by controllers.
	// In most cases, it will be required to look at controllers logs to properly triage those issues.
	InternalErrorReason = "InternalError"
)
