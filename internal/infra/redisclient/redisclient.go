package redisclient

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"github.com/rickferrdev/elysia-botbot/internal/infra/logger"
	"go.uber.org/fx"
)

type RedisClient struct {
	DB        *redis.Client
	lifecycle fx.Lifecycle
	env       *env.Env
	logger    *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
}

type FxArgs struct {
	fx.In
	Lifecycle fx.Lifecycle
	Env       *env.Env
	Logger    *slog.Logger
}

var (
	Provide = fx.Provide(New)
	Invoke  = fx.Invoke(func(redisclient *RedisClient) error {
		return redisclient.Start()
	})
)

func New(args FxArgs) (*RedisClient, error) {
	redisDB, err := strconv.Atoi(args.Env.RedisDB)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logger.RedisFailedParseIndexDB, err)
	}
	opts := redis.Options{
		Addr:     args.Env.RedisHost,
		Username: args.Env.RedisUser,
		Password: args.Env.RedisPass,
		DB:       redisDB,
		Protocol: 2,
	}

	rdc := redis.NewClient(&opts)
	client := RedisClient{
		lifecycle: args.Lifecycle,
		env:       args.Env,
		logger:    args.Logger,
		DB:        rdc,
	}

	return &client, nil
}

func (redisclient *RedisClient) Start() error {
	redisclient.lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if _, err := redisclient.DB.Ping(ctx).Result(); err != nil {
				return fmt.Errorf("%s: %w", logger.RedisFailedPingDB, err)
			}
			return nil
		},
		OnStop: func(fctx context.Context) error {
			if redisclient.cancel != nil {
				redisclient.cancel()
			}
			if err := redisclient.DB.Close(); err != nil {
				return fmt.Errorf("%s: %w", logger.RedisFailedCloseConnection, err)
			}
			return nil
		},
	})

	return nil
}
