package events

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/events/autoreactions"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"events",
	autoreactions.Invoke,
)
