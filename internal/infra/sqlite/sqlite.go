package sqlite

import (
	"database/sql"
	"log/slog"

	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"go.uber.org/fx"
)

var Provide = fx.Provide(New)
var Invoke = fx.Invoke(New)

type SqliteClient struct {
	DB *bun.DB
}

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    *slog.Logger
	Env       *env.Env
}

func New(params Params) (*SqliteClient, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, "file:local.db?cache=shared&mode=rwc")
	if err != nil {
		panic(err)
	}

	sqldb.SetMaxOpenConns(1)

	database := bun.NewDB(sqldb, sqlitedialect.New())

	if err := sqldb.Ping(); err != nil {
		return nil, err
	}

	sqliteClient := &SqliteClient{database}
	return sqliteClient, nil
}
