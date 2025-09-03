package main

import (
	cfg "github.com/conductorone/baton-incident-io/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("incident-io", cfg.Config)
}
