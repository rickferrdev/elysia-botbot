package autoreactions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/logger"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbautoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/platform/cache"
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
		cache         cache.Interface
		autoreactions dbautoreactions.Interface
	}

	FxArgs struct {
		fx.In

		Logger          *slog.Logger
		Cache           cache.Interface
		AutoReactionsDB dbautoreactions.Interface
	}
)

type Interface interface {
	ObtainAll(ctx context.Context, guildID string) ([]domain.AutoReaction, error)
	ObtainOne(ctx context.Context, guildID, channelID string) (*domain.AutoReaction, error)
	CreateOne(ctx context.Context, autoreaction domain.AutoReaction) (*domain.AutoReaction, error)
	UpdateOne(ctx context.Context, autoreaction domain.AutoReaction) (*domain.AutoReaction, error)
	DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReaction, error)
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

func (service *Service) CreateOne(ctx context.Context, autoreaction domain.AutoReaction) (*domain.AutoReaction, error) {
	data, err := service.autoreactions.CreateOne(ctx, autoreaction)
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
			logger.AutoReactionFailedInvalidateCache,
			slog.String("operation", "create"),
			slog.String("guild_id", autoreaction.GuildID),
			slog.String("channel_id", autoreaction.ChannelID),
			slog.Any("error", err),
		)
	}
	return data, nil
}

func (service *Service) ObtainOne(ctx context.Context, guildID, channelID string) (*domain.AutoReaction, error) {
	var cached domain.AutoReaction
	if found, err := service.cache.GetWithJson(
		ctx,
		service.getCacheKey(guildID, channelID),
		&cached,
	); err == nil && found {
		return &cached, nil
	}

	data, err := service.autoreactions.GetOne(ctx, guildID, channelID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrAutoReactionNotFound, err)
	}

	err = service.cache.SetWithJson(
		ctx,
		service.getCacheKey(guildID, channelID),
		data,
		autoreactionsCacheTTL,
	)
	if err != nil {
		service.logger.Error(
			logger.AutoReactionFailedCache,
			slog.String("guild_id", guildID),
			slog.String("channel_id", channelID),
			slog.Any("error", err),
		)
	}

	return data, nil
}

func (service *Service) UpdateOne(ctx context.Context, autoreaction domain.AutoReaction) (*domain.AutoReaction, error) {
	data, err := service.autoreactions.UpdateOne(ctx, autoreaction)
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
			logger.AutoReactionFailedInvalidateCache,
			slog.String("operation", "update"),
			slog.String("guild_id", autoreaction.GuildID),
			slog.String("channel_id", autoreaction.ChannelID),
			slog.Any("error", err),
		)
	}
	return data, nil
}

func (service *Service) DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReaction, error) {
	data, err := service.autoreactions.DeleteOne(ctx, guildID, channelID)
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
			logger.AutoReactionFailedInvalidateCache,
			slog.String("operation", "delete"),
			slog.String("guild_id", guildID),
			slog.String("channel_id", channelID),
			slog.Any("error", err),
		)
	}
	return data, nil
}

func (service *Service) ObtainAll(ctx context.Context, guildID string) ([]domain.AutoReaction, error) {
	var cached []domain.AutoReaction
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
			logger.AutoReactionFailedCacheAll,
			slog.String("guild_id", guildID),
			slog.Any("error", err),
		)
	}
	return result, nil
}
