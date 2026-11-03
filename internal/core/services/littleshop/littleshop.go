package littleshop

import (
	"context"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbslipper"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtomatoes"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

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

func New(FxArgs FxArgs) *Service {
	service := Service{
		dbtomatoes: FxArgs.TomatoesDB,
		dbslipper:  FxArgs.SlipperDB,
	}

	return &service
}

func (service *Service) PurchaseTomato(ctx context.Context, guildID, userID string, quantity int) (int, error) {
	slipper, err := service.dbslipper.GetInventory(ctx, guildID, userID)
	if err != nil {
		return 0, InternalError
	}

	cost := quantity * domain.TomatoesCost

	if slipper < cost {
		return 0, InsufficientBalanceError
	}

	balance, err := service.dbtomatoes.Increment(ctx, guildID, userID, quantity)
	if err != nil {
		return 0, InternalError
	}

	_, err = service.dbslipper.DecrementWith(ctx, guildID, userID, cost)
	if err != nil {
		return 0, InternalError
	}

	return balance, nil
}
