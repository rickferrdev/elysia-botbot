package repositories

import (
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbautoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbservers"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbslipper"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtemplates"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtomatoes"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"repositories",

	dbservers.Provide,
	dbtemplates.Provide,
	dbtomatoes.Provide,
	dbslipper.Provide,
	dbautoreactions.Provide,
)
