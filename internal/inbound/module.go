package inbound

import (
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"inbound",
	interactions.Module,
)
