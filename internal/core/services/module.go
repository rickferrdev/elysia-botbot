package services

import (
	"github.com/rickferrdev/elysia-botbot/internal/core/services/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/littleshop"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/server"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/slipper"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/templates"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/tomatoes"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"services",

	templates.Provide,
	server.Provide,
	interactionflow.Provide,
	autoreactions.Provide,
	slipper.Provide,
	tomatoes.Provide,
	littleshop.Provide,
	responder.Provide,
)
