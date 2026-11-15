package littleshop

import (
	"context"
	"errors"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbslipper"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtomatoes"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

var (
	ErrTomatoPurchaseFailed       = errors.New("failed to purchase tomatoes")
	ErrInsufficientSlipperBalance = errors.New("insufficient slipper balance to purchase tomatoes")
)

type (
	Service struct {
		dbtomatoes dbtomatoes.Interface
		dbslipper  dbslipper.Interface
	}

	FxArgs struct {
		fx.In

		TomatoesDB dbtomatoes.Interface
		SlipperDB  dbslipper.Interface
	}
)

type Interface interface {
	PurchaseTomato(ctx context.Context, guildID, userID string, quantity int) (int, error)
}

func New(args FxArgs) *Service {
	service := Service{
		dbtomatoes: args.TomatoesDB,
		dbslipper:  args.SlipperDB,
	}

	return &service
}

func (service *Service) PurchaseTomato(ctx context.Context, guildID, userID string, quantity int) (int, error) {
	slipper, err := service.dbslipper.GetInventory(ctx, guildID, userID)
	if err != nil {
		return 0, ErrTomatoPurchaseFailed
	}

	cost := quantity * domain.TomatoesCost

	if slipper < cost {
		return 0, ErrInsufficientSlipperBalance
	}

	balance, err := service.dbtomatoes.Increment(ctx, guildID, userID, quantity)
	if err != nil {
		return 0, ErrTomatoPurchaseFailed
	}

	_, err = service.dbslipper.DecrementWith(ctx, guildID, userID, cost)
	if err != nil {
		return 0, ErrTomatoPurchaseFailed
	}

	return balance, nil
}
