package services

import (
	"github.com/rickferrdev/elysia-botbot/internal/core/services/funnies/kwaii"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/funnies/slipper"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/funnies/tomatoes"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/server"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/server/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/server/templates"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/store/littleshop"
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
	kwaii.Provide,
)
