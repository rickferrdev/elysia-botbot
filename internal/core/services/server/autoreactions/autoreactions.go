package autoreactions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbautoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/platform/platcache"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

var (
	ErrAutoReactionNotFound      = errors.New("auto-reaction not found")
	ErrAutoReactionAlreadyExists = errors.New("auto-reaction already exists")
)

type (
	Service struct {
		logger        *slog.Logger
		cache         platcache.Interface
		autoreactions dbautoreactions.Interface
	}

	FxArgs struct {
		fx.In

		Logger          *slog.Logger
		Cache           platcache.Interface
		AutoReactionsDB dbautoreactions.Interface
	}
)

type Interface interface {
	CreateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error)
	ObtainOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error)
	ObtainAll(ctx context.Context, guildID string) ([]domain.AutoReactions, error)
	UpdateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error)
	DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error)
}

func New(args FxArgs) *Service {
	service := Service{
		logger:        args.Logger,
		cache:         args.Cache,
		autoreactions: args.AutoReactionsDB,
	}

	return &service
}

const autoreactionsCacheTTL = time.Minute * 10

func (service *Service) getCacheKey(guildID, channelID string) string {
	return "autoreactions:" + guildID + ":" + channelID
}

func (service *Service) getAllCacheKey(guildID string) string {
	return "autoreactions:guild:" + guildID
}

func (service *Service) CreateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error) {
	result, err := service.autoreactions.CreateOne(ctx, autoreaction)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlConflict) {
			return nil, fmt.Errorf("%w: %w", ErrAutoReactionAlreadyExists, err)
		}
		return nil, err
	}

	err = service.cache.Del(
		ctx,
		service.getCacheKey(autoreaction.GuildID, autoreaction.ChannelID),
		service.getAllCacheKey(autoreaction.GuildID),
	)
	if err != nil {
		service.logger.Error(
			"failed to invalidate auto-reaction cache",
			slog.String("operation", "create"),
			slog.String("guild_id", autoreaction.GuildID),
			slog.String("channel_id", autoreaction.ChannelID),
			slog.Any("error", err),
		)
	}
	return result, nil
}

func (service *Service) ObtainOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error) {
	var cached domain.AutoReactions
	if found, err := service.cache.GetWithJson(
		ctx,
		service.getCacheKey(guildID, channelID),
		&cached,
	); err == nil && found {
		return &cached, nil
	}

	result, err := service.autoreactions.GetOne(ctx, guildID, channelID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrAutoReactionNotFound, err)
	}

	err = service.cache.SetWithJson(
		ctx,
		service.getCacheKey(guildID, channelID),
		result,
		autoreactionsCacheTTL,
	)
	if err != nil {
		service.logger.Error(
			"failed to cache auto-reaction",
			slog.String("guild_id", guildID),
			slog.String("channel_id", channelID),
			slog.Any("error", err),
		)
	}

	return result, nil
}

func (service *Service) UpdateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error) {
	result, err := service.autoreactions.UpdateOne(ctx, autoreaction)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrAutoReactionNotFound, err)
		}
		return nil, err
	}

	err = service.cache.Del(
		ctx,
		service.getCacheKey(autoreaction.GuildID, autoreaction.ChannelID),
		service.getAllCacheKey(autoreaction.GuildID),
	)
	if err != nil {
		service.logger.Error(
			"failed to invalidate auto-reaction cache",
			slog.String("operation", "update"),
			slog.String("guild_id", autoreaction.GuildID),
			slog.String("channel_id", autoreaction.ChannelID),
			slog.Any("error", err),
		)
	}
	return result, nil
}

func (service *Service) DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error) {
	result, err := service.autoreactions.DeleteOne(ctx, guildID, channelID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrAutoReactionNotFound, err)
	}

	err = service.cache.Del(
		ctx,
		service.getCacheKey(guildID, channelID),
		service.getAllCacheKey(guildID),
	)
	if err != nil {
		service.logger.Error(
			"failed to invalidate auto-reaction cache",
			slog.String("operation", "delete"),
			slog.String("guild_id", guildID),
			slog.String("channel_id", channelID),
			slog.Any("error", err),
		)
	}
	return result, nil
}

func (service *Service) ObtainAll(ctx context.Context, guildID string) ([]domain.AutoReactions, error) {
	var cached []domain.AutoReactions
	if found, err := service.cache.GetWithJson(
		ctx,
		service.getAllCacheKey(guildID),
		&cached,
	); err == nil && found {
		return cached, nil
	}

	result, err := service.autoreactions.GetAll(ctx, guildID)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, ErrAutoReactionNotFound
		}

		return nil, err
	}

	err = service.cache.SetWithJson(
		ctx,
		service.getAllCacheKey(guildID),
		result,
		autoreactionsCacheTTL,
	)
	if err != nil {
		service.logger.Error(
			"failed to cache auto-reactions",
			slog.String("guild_id", guildID),
			slog.Any("error", err),
		)
	}
	return result, nil
}
