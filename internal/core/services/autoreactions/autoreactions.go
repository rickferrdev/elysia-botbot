package autoreactions

import (
	"context"
	"fmt"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/autoreactions"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		autoreactions autoreactions.Interface
	}

	FxArgs struct {
		fx.In

		AutoReactionsDB autoreactions.Interface
		DiscordClient *discordclient.DiscordClient
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
		autoreactions: FxArgs.AutoReactionsDB,
	}

	return &service
}

func (service *Service) CreateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error) {
	result, err := service.autoreactions.CreateOne(ctx, autoreaction)
	if core.IsCode(err, repoerrors.CodeSqlConflict) {
		return nil, fmt.Errorf("%w: %w", ErrConflict, err)
	}

	return result, err
}

func (service *Service) ObtainOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error) {
	result, err := service.autoreactions.GetOne(ctx, guildID, channelID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	return result, err
}

func (service *Service) UpdateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error) {
	result, err := service.autoreactions.UpdateOne(ctx, autoreaction)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	return result, err
}

func (service *Service) DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error) {
	result, err := service.autoreactions.DeleteOne(ctx, guildID, channelID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	return result, err
}

func (service *Service) ObtainAll(ctx context.Context, guildID string) ([]domain.AutoReactions, error) {
	result, err := service.autoreactions.GetAll(ctx, guildID)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return result, nil
}
