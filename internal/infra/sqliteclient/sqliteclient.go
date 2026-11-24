package sqliteclient

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"github.com/rickferrdev/elysia-botbot/internal/infra/logger"
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
		args.Logger.Warn(logger.SQLiteFailedConnectionAttempt, slog.Any("error", err))
		connection, err = sql.Open(sqliteshim.ShimName, "file:database?mode=memory&cache=shared")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logger.SQLiteFailedConnectionMemoryMode, err)
		}
	}

	if connection == nil {
		return nil, fmt.Errorf("%s: %w", logger.SQLiteFailedConnectionNil, err)
	}

	connection.SetMaxOpenConns(1)
	database := bun.NewDB(connection, sqlitedialect.New())

	args.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()
			if err := database.PingContext(ctx); err != nil {
				return fmt.Errorf("%s: %w", logger.SQLiteFailedPingDB, err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := database.Close(); err != nil {
				return fmt.Errorf("%s: %w", logger.SQLiteFailedCloseConnection, err)
			}
			return nil
		},
	})

	client := &SqliteClient{database}
	return client, nil
}
