package commands

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/embeds"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/ping"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/server"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/status"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"commands",
	autoreactions.Invoke,
	server.Invoke,
	status.Invoke,
	ping.Invoke,
	embeds.Invoke,
)
