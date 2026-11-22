package servers

import (
	"context"
	"errors"
	"fmt"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbservers"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

var (
	ErrServerConfigurationNotFound      = errors.New("server configuration not found")
	ErrServerConfigurationAlreadyExists = errors.New("server configuration already exists")
)

type (
	Service struct {
		storage dbservers.Interface
	}

	FxArgs struct {
		fx.In

		ServerStorage dbservers.Interface
	}
)

type Interface interface {
	CreateServerWithConfig(ctx context.Context, server domain.Server) error
	ObtainServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error)
}

func New(args FxArgs) *Service {
	service := Service{
		storage: args.ServerStorage,
	}

	return &service
}

func (service *Service) CreateServerWithConfig(ctx context.Context, server domain.Server) error {
	_, err := service.storage.CreateServerWithConfig(ctx, server)
	if core.IsCode(err, repoerrors.CodeSqlConflict) {
		return fmt.Errorf("%w: %w", ErrServerConfigurationAlreadyExists, err)
	}

	return err
}

func (service *Service) ObtainServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error) {
	result, err := service.storage.GetServerWithConfig(ctx, guildID)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrServerConfigurationNotFound, err)
	}

	return result, err
}
