package events

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/events/funnies/kwaii"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/events/server/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/events/witness"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"events",
	autoreactions.Invoke,
	kwaii.Invoke,
	witness.Invoke,
)
