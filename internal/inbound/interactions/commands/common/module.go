package common

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/common/bestchoice"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/common/humiliator"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/common/slipper"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/commands/common/tomato"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"common",
	slipper.Invoke,
	tomato.Invoke,
	bestchoice.Invoke,
	humiliator.Invoke,
)
