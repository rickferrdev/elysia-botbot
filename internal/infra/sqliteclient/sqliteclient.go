package sqliteclient

import (
	"database/sql"
	"fmt"
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

type FxArgs struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    *slog.Logger
	Env       *env.Env
}

func New(args FxArgs) (*SqliteClient, error) {
	dataSource := fmt.Sprintf("file:%v?cache=shared&mode=rwc", args.Env.SqliteDB)
	sqldb, err := sql.Open(sqliteshim.ShimName, dataSource)
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
