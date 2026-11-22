package platform

import (
	"github.com/rickferrdev/elysia-botbot/internal/platform/cache"
	"github.com/rickferrdev/elysia-botbot/internal/platform/genai"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"platform",

	genai.Provide,
	cache.Provide,
)
