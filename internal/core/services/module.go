package services

import (
	"github.com/rickferrdev/elysia-botbot/internal/core/services/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/embeds"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/funslipper"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/server"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"services",

	embeds.Provide,
	server.Provide,
	interactionflow.Provide,
	autoreactions.Provide,
	responder.Provide,
	funslipper.Provide,
)
