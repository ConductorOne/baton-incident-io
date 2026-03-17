package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-incident-io/pkg/config"
	"github.com/conductorone/baton-incident-io/pkg/connector"
	sdkconfig "github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
	"github.com/conductorone/baton-sdk/pkg/field"

	"github.com/conductorone/baton-sdk/pkg/cli"
)

var version = "dev"

func main() {
	ctx := context.Background()

	sdkconfig.RunConnector(
		ctx,
		"baton-incident-io",
		version,
		config.Config,
		getConnector,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilderV2(&connector.IncidentIo{}),
	)
}

func getConnector(ctx context.Context, cfg *config.IncidentIo, opts *cli.ConnectorOpts) (connectorbuilder.ConnectorBuilderV2, []connectorbuilder.Opt, error) {
	if err := field.Validate(config.Config, cfg); err != nil {
		return nil, nil, err
	}

	accessToken := cfg.GetString("token")

	if accessToken == "" {
		return nil, nil, fmt.Errorf("missing access token")
	}

	cb, err := connector.New(ctx, accessToken)
	if err != nil {
		return nil, nil, err
	}

	return cb, nil, nil
}
