package services

import (
	"github.com/rickferrdev/elysia-botbot/internal/core/services/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/effects"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/servers"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/templates"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"services",

	templates.Provide,
	servers.Provide,
	interactionflow.Provide,
	autoreactions.Provide,
	responder.Provide,
	effects.Provide,
)
