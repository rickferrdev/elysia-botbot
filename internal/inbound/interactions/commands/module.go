package commands

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/embeds"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/server"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"commands",
	autoreactions.Invoke,
	server.Invoke,
	embeds.Invoke,
)
