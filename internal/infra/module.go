package infra

import (
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"github.com/rickferrdev/elysia-botbot/internal/infra/logger"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqlite"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"infra",
	discord.Provide,
	discord.Invoke,
	logger.Provide,

	sqlite.Provide,
	sqlite.Invoke,
)
