package platform

import (
	"github.com/rickferrdev/elysia-botbot/internal/platform/platcache"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"platform",
	platcache.Provide,
)
