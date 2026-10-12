package repositories

import (
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/embeds"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/server"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"repositories",
	server.Provide,
	autoreactions.Provide,
	embeds.Provide,
)
