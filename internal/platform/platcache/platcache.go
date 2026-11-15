package platcache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/infra/redisclient"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Platform struct {
		client *redisclient.RedisClient
	}

	FxArgs struct {
		fx.In

		RedisClient *redisclient.RedisClient
	}

	Interface interface {
		Set(ctx context.Context, key string, value any, expiration time.Duration) error
		Get(ctx context.Context, key string) (bool, []byte, error)
		Del(ctx context.Context, keys ...string) error
		GetWithJson(ctx context.Context, key string, dest any) (bool, error)
		SetWithJson(ctx context.Context, key string, value any, expiration time.Duration) error
	}
)

func New(args FxArgs) *Platform {
	redis := Platform{
		client: args.RedisClient,
	}

	return &redis
}

func (platform *Platform) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return platform.client.DB.Set(ctx, key, value, expiration).Err()
}

func (platform *Platform) Get(ctx context.Context, key string) (bool, []byte, error) {
	data, err := platform.client.DB.Get(ctx, key).Bytes()
	if err != nil {
		return false, []byte{}, err
	}

	return true, data, nil
}

func (platform *Platform) Del(ctx context.Context, keys ...string) error {
	return platform.client.DB.Del(ctx, keys...).Err()
}

func (platform *Platform) GetWithJson(ctx context.Context, key string, dest any) (bool, error) {
	data, err := platform.client.DB.Get(ctx, key).Bytes()
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}

	return true, nil
}

func (platform *Platform) SetWithJson(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return platform.client.DB.Set(
		ctx,
		key,
		data,
		expiration,
	).Err()
}
