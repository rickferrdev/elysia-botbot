package commands

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/common"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/littleshop"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/server"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/templates"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"commands",

	autoreactions.Invoke,
	server.Invoke,
	templates.Invoke,
	littleshop.Invoke,
	common.Module,
)
