package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	tokenField = field.StringField(
		"token",
		field.WithDescription("token"),
		field.WithRequired(true),
		field.WithIsSecret(true),
		field.WithDisplayName("Personal access token"),
	)
)

var configFields = []field.SchemaField{tokenField}

// Config represents the configuration for the baton-incident-io connector.
var Config = field.NewConfiguration(
	configFields,
	field.WithConnectorDisplayName("Incident.io"),
	field.WithHelpUrl("/docs/baton/incident-io"),
	field.WithIconUrl("/static/app-icons/incident-io.svg"),
)
