package repositories

import (
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbautoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbeffects"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbservers"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtemplates"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"repositories",

	dbservers.Provide,
	dbtemplates.Provide,
	dbautoreactions.Provide,
	dbeffects.Provide,
)
