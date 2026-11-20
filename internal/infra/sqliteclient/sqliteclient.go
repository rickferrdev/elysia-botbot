package sqliteclient

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"go.uber.org/fx"
)

var Provide = fx.Provide(New)

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
	var connection *sql.DB
	dataSource := fmt.Sprintf("file:%v?cache=shared&mode=rwc", args.Env.SqliteDB)
	connection, err := sql.Open(sqliteshim.ShimName, dataSource)
	if err != nil {
		args.Logger.Warn("failed to connect to the database; in-memory mode will be used", slog.Any("error", err))
		connection, err = sql.Open(sqliteshim.ShimName, "file:database?mode=memory&cache=shared")
		if err != nil {
			args.Logger.Error("failed to connect to memory mode", slog.Any("error", err))
			return nil, fmt.Errorf("failed to connect to memory mode: %w", err)
		}
	}

	if connection == nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	connection.SetMaxOpenConns(1)
	database := bun.NewDB(connection, sqlitedialect.New())

	args.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()
			return database.PingContext(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return database.Close()
		},
	})

	client := &SqliteClient{database}
	return client, nil
}
