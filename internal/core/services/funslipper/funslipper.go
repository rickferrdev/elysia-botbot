package funslipper

import (
	"context"
	"fmt"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/funslipper"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		funslipper funslipper.Interface
	}

	FxArgs struct {
		fx.In
		FunSlipper funslipper.Interface
	}

	Interface interface {
		IncrementCount(ctx context.Context, guildID, userID string) (*domain.FunSlipper, error)
	}
)

func New(args FxArgs) *Service {
	service := Service{
		funslipper: args.FunSlipper,
	}

	return &service
}

func (service *Service) IncrementCount(ctx context.Context, guildID, userID string) (*domain.FunSlipper, error) {
	funslipper := domain.FunSlipper{
		GuildID: guildID,
		UserID:  userID,
	}
	data, err := service.funslipper.IncrementCount(ctx, funslipper)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", err, ErrConflict)
	}
	return data, nil
}
