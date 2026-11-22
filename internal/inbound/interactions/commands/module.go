package commands

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/effects"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/funnies"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/servers"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/templates"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"commands",

	effects.Invoke,
	servers.Invoke,
	autoreactions.Invoke,
	templates.Invoke,
	funnies.Module,
)
