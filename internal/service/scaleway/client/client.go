package client

import (
	"fmt"
	"slices"

	"github.com/scaleway/cluster-api-provider-scaleway/internal/version"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	createdByTag         = "created-by=cluster-api-provider-scaleway"
	createdByDescription = "Created by cluster-api-provider-scaleway"
)

var userAgent = "cluster-api-provider-scaleway/" + version.Version

// Client is a wrapper over scaleway-sdk-go to access Scaleway Product APIs in
// a specific region and project.
type Client struct {
	// Client scope.
	projectID string
	region    scw.Region

	// Product APIs
	vpc         *vpc.API
	vpcgw       *vpcgw.API
	lb          *lb.ZonedAPI
	domain      *domain.API
	instance    *instance.API
	block       *block.API
	marketplace *marketplace.API
	ipam        *ipam.API
}

// New returns a new Scaleway client based on the provided region and secretData.
// The secret data must contain a default projectID and credentials.
func New(region scw.Region, projectID string, secretData map[string][]byte) (*Client, error) {
	accessKey := string(secretData[scw.ScwAccessKeyEnv])
	if accessKey == "" {
		return nil, fmt.Errorf("field %s is missing in secret", scw.ScwAccessKeyEnv)
	}

	secretKey := string(secretData[scw.ScwSecretKeyEnv])
	if secretKey == "" {
		return nil, fmt.Errorf("field %s is missing in secret", scw.ScwSecretKeyEnv)
	}

	opts := []scw.ClientOption{
		scw.WithAuth(accessKey, secretKey),
		scw.WithDefaultProjectID(projectID),
		scw.WithDefaultRegion(region),
		scw.WithUserAgent(userAgent),
	}

	if apiURL := string(secretData[scw.ScwAPIURLEnv]); apiURL != "" {
		opts = append(opts, scw.WithAPIURL(apiURL))
	}

	client, err := scw.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create scaleway-sdk-go client: %w", err)
	}

	return &Client{
		projectID:   projectID,
		region:      region,
		vpc:         vpc.NewAPI(client),
		vpcgw:       vpcgw.NewAPI(client),
		lb:          lb.NewZonedAPI(client),
		domain:      domain.NewAPI(client),
		instance:    instance.NewAPI(client),
		block:       block.NewAPI(client),
		marketplace: marketplace.NewAPI(client),
		ipam:        ipam.NewAPI(client),
	}, nil
}

func matchTags(tags []string, wantedTags []string) bool {
	for _, tag := range wantedTags {
		if !slices.Contains(tags, tag) {
			return false
		}
	}

	return true
}

func validateTags(tags []string) error {
	if len(tags) == 0 {
		return fmt.Errorf("tags cannot be empty")
	}

	return nil
}
