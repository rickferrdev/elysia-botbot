package outbound

import (
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"outbound",
	repositories.Module,
)
