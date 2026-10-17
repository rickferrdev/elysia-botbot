package infra

import (
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"github.com/rickferrdev/elysia-botbot/internal/infra/logger"
	"github.com/rickferrdev/elysia-botbot/internal/infra/redisclient"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"infra",

	discordclient.Provide,
	discordclient.Invoke,

	sqliteclient.Provide,
	sqliteclient.Invoke,

	redisclient.Provide,
	redisclient.Invoke,
	logger.Provide,
)
