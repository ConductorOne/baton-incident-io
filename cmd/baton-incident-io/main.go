package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-incident-io/pkg/config"
	"github.com/conductorone/baton-incident-io/pkg/connector"
	sdkconfig "github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := sdkconfig.DefineConfiguration(
		ctx,
		"baton-incident-io",
		getConnector,
		config.Config,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilder(&connector.Connector{}),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, cfg *config.IncidentIo) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	if err := field.Validate(config.Config, cfg); err != nil {
		return nil, err
	}

	accessToken := cfg.GetString("token")

	if accessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}

	cb, err := connector.New(ctx, accessToken)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	connector, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	return connector, nil
}
