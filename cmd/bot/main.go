// Modules are grouped by responsibility.
//
// config: application configuration and environment variables.
// infra: framework-level infrastructure, such as HTTP server and logger.
// outbound: adapters that communicate with external systems.
// services: application use cases / business services.
// inbound: adapters that receive input, such as HTTP handlers.
package main

import (
	"github.com/rickferrdev/elysia-botbot/internal/config"
	"github.com/rickferrdev/elysia-botbot/internal/core/services"
	"github.com/rickferrdev/elysia-botbot/internal/inbound"
	"github.com/rickferrdev/elysia-botbot/internal/infra"
	"github.com/rickferrdev/elysia-botbot/internal/outbound"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		infra.Module,
		outbound.Module,
		services.Module,
		inbound.Module,
	).Run()
}
