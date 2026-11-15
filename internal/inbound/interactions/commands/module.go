package commands

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/funnies"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/server"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/server/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/server/templates"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/store/littleshop"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"commands",

	server.Invoke,
	autoreactions.Invoke,
	templates.Invoke,
	littleshop.Invoke,
	funnies.Module,
)
