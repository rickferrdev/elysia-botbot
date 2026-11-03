package slipper

import (
	"context"

	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbslipper"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		dbslipper dbslipper.Interface
	}

	FxArgs struct {
		fx.In

		SlipperDB dbslipper.Interface
	}

	Interface interface {
		ThrowSlipper(ctx context.Context, guildID, userID string) (int, error)
	}
)

func New(args FxArgs) *Service {
	service := Service{
		dbslipper: args.SlipperDB,
	}

	return &service
}

func (service *Service) ThrowSlipper(ctx context.Context, guildID, userID string) (int, error) {
	inventory, err := service.dbslipper.IncrementWith(ctx, guildID, userID, 1)
	if err != nil {
		return 0, InternalError
	}

	return inventory, nil
}
