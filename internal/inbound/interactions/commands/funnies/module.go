package funnies

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/funnies/bestchoice"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/funnies/humiliator"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"funnies",

	bestchoice.Invoke,
	humiliator.Invoke,
)
