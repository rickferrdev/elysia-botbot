package tomatoes

import (
	"context"
	"errors"
	"fmt"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtomatoes"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

var (
	ErrTomatoThrowFailed         = errors.New("failed to throw a tomato")
	ErrInsufficientTomatoBalance = errors.New("insufficient tomato balance")
)

type (
	Service struct {
		dbtomatoes dbtomatoes.Interface
	}

	FxArgs struct {
		fx.In

		TomatoesDB dbtomatoes.Interface
	}

	Interface interface {
		ThrowTomato(ctx context.Context, guildID, userID string) (int, error)
	}
)

func New(args FxArgs) *Service {
	service := Service{
		dbtomatoes: args.TomatoesDB,
	}

	return &service
}

func (service *Service) ThrowTomato(ctx context.Context, guildID, userID string) (int, error) {
	balance, err := service.dbtomatoes.GetInventory(ctx, guildID, userID)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return 0, ErrInsufficientTomatoBalance
		}
		return 0, fmt.Errorf("%w: %w", ErrTomatoThrowFailed, err)
	}

	if balance == 0 {
		return 0, ErrInsufficientTomatoBalance
	}

	balance, err = service.dbtomatoes.Decrement(ctx, guildID, userID, 1)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrTomatoThrowFailed, err)
	}

	return balance, nil
}
