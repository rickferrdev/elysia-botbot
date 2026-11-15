package funnies

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/funnies/bestchoice"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/funnies/humiliator"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/funnies/kwaii"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/funnies/slipper"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/funnies/tomato"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"funnies",
	slipper.Invoke,
	tomato.Invoke,
	bestchoice.Invoke,
	humiliator.Invoke,
	kwaii.Invoke,
)
