package redisclient

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rickferrdev/elysia-botbot/internal/config/env"
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

var Provide = fx.Provide(New)
var Invoke = fx.Invoke(func(redisclient *RedisClient) error {
	return redisclient.Start()
})

func New(args FxArgs) (*RedisClient, error) {
	redisDB, err := strconv.Atoi(args.Env.RedisDB)
	if err != nil {
		return nil, err
	}
	opts := redis.Options{
		Addr:       args.Env.RedisHost,
		Username:   args.Env.RedisUser,
		Password:   args.Env.RedisPass,
		DB:         redisDB,
		ClientName: "elysia-botbot",
		Protocol:   2,
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
				return err
			}
			return nil
		},
		OnStop: func(fctx context.Context) error {
			if redisclient.cancel != nil {
				redisclient.cancel()
			}
			return redisclient.DB.Close()
		},
	})

	return nil
}
