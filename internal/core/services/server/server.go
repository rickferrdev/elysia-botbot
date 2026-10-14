package server

import (
	"context"
	"fmt"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/server"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		storage server.Interface
	}
	Params struct {
		fx.In
		ServerStorage server.Interface
	}
)

type Interface interface {
	CreateServerWithConfig(ctx context.Context, server domain.Server) error
	ObtainServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error)
}

func New(params Params) *Service {
	service := Service{
		storage: params.ServerStorage,
	}

	return &service
}

func (service *Service) CreateServerWithConfig(ctx context.Context, server domain.Server) error {
	err := service.storage.CreateServerWithConfig(ctx, server)
	if core.IsCode(err, repoerrors.CodeSqlConflict) {
		return fmt.Errorf("%w: %w", ErrConflict, err)
	}
	return err
}
func (service *Service) ObtainServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error) {
	result, err := service.storage.GetServerWithConfig(ctx, guildID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return result, err
}
