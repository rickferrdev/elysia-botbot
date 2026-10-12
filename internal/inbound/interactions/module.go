package interactions

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/events"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"interactions",
	commands.Module,
	events.Module,
)
