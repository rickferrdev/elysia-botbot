package autoreactions

import (
	"context"
	"fmt"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/redisclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/autoreactions"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		redis         *redisclient.RedisClient
		autoreactions autoreactions.Interface
	}

	FxArgs struct {
		fx.In

		RedisClient     *redisclient.RedisClient
		AutoReactionsDB autoreactions.Interface
	}
)

type Interface interface {
	CreateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error)
	ObtainOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error)
	ObtainAll(ctx context.Context, guildID string) ([]domain.AutoReactions, error)
	UpdateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error)
	DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error)
}

func New(FxArgs FxArgs) *Service {
	service := Service{
		redis:         FxArgs.RedisClient,
		autoreactions: FxArgs.AutoReactionsDB,
	}

	return &service
}

func (service *Service) CreateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error) {
	result, err := service.autoreactions.CreateOne(ctx, autoreaction)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlConflict) {
			return nil, fmt.Errorf("%w: %w", ErrConflict, err)
		}
		return nil, err
	}

	_ = service.delCache(ctx, autoreaction.GuildID, autoreaction.ChannelID)

	return result, nil
}

func (service *Service) ObtainOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error) {
	var cached domain.AutoReactions
	if found, err := service.getCache(
		ctx,
		service.getCacheKey(guildID, channelID),
		&cached,
	); err == nil && found {
		fmt.Printf("get cache do ObtainOne: %v\n", cached)
		return &cached, nil
	}

	fmt.Println("no cache found in ObtainOne")

	result, err := service.autoreactions.GetOne(ctx, guildID, channelID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	err = service.setCache(
		ctx,
		service.getCacheKey(guildID, channelID),
		result,
	)
	fmt.Printf("set cache do ObtainOne: %v\n", err)

	return result, err
}

func (service *Service) UpdateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error) {
	result, err := service.autoreactions.UpdateOne(ctx, autoreaction)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
		}
		return nil, err
	}

	_ = service.delCache(ctx, autoreaction.GuildID, autoreaction.ChannelID)

	return result, nil
}

func (service *Service) DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error) {
	result, err := service.autoreactions.DeleteOne(ctx, guildID, channelID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	_ = service.delCache(ctx, guildID, channelID)
	fmt.Printf("del cache do DeleteOne: %v\n", err)

	return result, err
}

func (service *Service) ObtainAll(ctx context.Context, guildID string) ([]domain.AutoReactions, error) {
	var cached []domain.AutoReactions
	if found, err := service.getCache(
		ctx,
		service.getAllCacheKey(guildID),
		&cached,
	); err == nil && found {
		fmt.Printf("get cache do ObtainAll: %v\n", cached)
		return cached, nil
	}

	fmt.Println("no cache found in ObtainOne")

	result, err := service.autoreactions.GetAll(ctx, guildID)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	err = service.setCache(
		ctx,
		service.getAllCacheKey(guildID),
		result,
	)
	fmt.Printf("set cache do ObtainAll: %v\n", err)

	return result, nil
}
