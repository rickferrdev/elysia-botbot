package infra

import (
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"github.com/rickferrdev/elysia-botbot/internal/infra/logger"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"infra",
	discordclient.Provide,
	discordclient.Invoke,
	logger.Provide,

	sqliteclient.Provide,
	sqliteclient.Invoke,
)
