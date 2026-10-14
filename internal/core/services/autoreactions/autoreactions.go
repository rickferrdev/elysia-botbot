package autoreactions

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/autoreactions"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		responder responder.Interface
		database  autoreactions.Interface
		flow      interactionflow.Interface
		client    *discord.DiscordClient
		logger    *slog.Logger
	}

	Params struct {
		fx.In
		Responder              responder.Interface
		InteractionFlow        interactionflow.Interface
		AutoReactionsInterface autoreactions.Interface
		Logger                 *slog.Logger
		DiscordClient          *discord.DiscordClient
	}
)

type Interface interface {
	CreateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error)
	ObtainOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error)
	UpdateOne(ctx context.Context, guildID, channelID, emoji string) (*domain.AutoReactions, error)
	DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error)
}

func New(params Params) *Service {
	service := Service{
		flow:      params.InteractionFlow,
		responder: params.Responder,
		client:    params.DiscordClient,
		logger:    params.Logger,
		database:  params.AutoReactionsInterface,
	}

	return &service
}

func (service *Service) CreateOne(ctx context.Context, autoreaction domain.AutoReactions) (*domain.AutoReactions, error) {
	result, err := service.database.CreateOne(ctx, autoreaction)
	if core.IsCode(err, repoerrors.CodeSqlConflict) {
		return nil, fmt.Errorf("%w: %w", ErrConflict, err)
	}
	return result, err
}

func (service *Service) ObtainOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error) {
	result, err := service.database.GetOne(ctx, guildID, channelID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return result, err
}

func (service *Service) UpdateOne(ctx context.Context, guildID, channelID, emoji string) (*domain.AutoReactions, error) {
	result, err := service.database.UpdateOne(ctx, guildID, channelID, emoji)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return result, err
}

func (service *Service) DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error) {
	result, err := service.database.DeleteOne(ctx, guildID, channelID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return result, err
}
