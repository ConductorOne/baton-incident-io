package connector

import (
	"context"
	"io"

	"github.com/conductorone/baton-incident-io/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type IncidentIo struct {
	apiClient       *client.APIClient
	syncBaseRoles   bool
	syncCustomRoles bool
}

// ResourceSyncers returns a ResourceSyncerV2 for each resource type that should be synced from the upstream service.
func (d *IncidentIo) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
		NewUserBuilder(d.apiClient, d.syncBaseRoles, d.syncCustomRoles),
		NewScheduleBuilder(d.apiClient),
		NewBaseRoleBuilder(d.apiClient),
		NewCustomRoleBuilder(d.apiClient),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *IncidentIo) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *IncidentIo) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Incidents.io connector",
		Description: "sync users and schedules from incidents.io",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *IncidentIo) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, accessToken string, opts *cli.ConnectorOpts) (*IncidentIo, error) {
	apiClient := client.NewClient(accessToken, nil)

	syncBaseRoles := opts.WillSyncResourceType(BaseRoleResourceTypeID)
	syncCustomRoles := opts.WillSyncResourceType(CustomRoleResourceTypeID)

	return &IncidentIo{
		apiClient:       apiClient,
		syncBaseRoles:   syncBaseRoles,
		syncCustomRoles: syncCustomRoles,
	}, nil
}
