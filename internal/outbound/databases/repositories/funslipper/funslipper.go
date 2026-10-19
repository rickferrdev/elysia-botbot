package funslipper

import (
	"context"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Database struct {
		sqlite *sqliteclient.SqliteClient
	}

	FxArgs struct {
		fx.In

		SqliteClient *sqliteclient.SqliteClient
	}
)

type Interface interface {
	IncrementCount(ctx context.Context, funslipper domain.FunSlipper) (*domain.FunSlipper, error)
}

func New(FxArgs FxArgs) (*Database, error) {
	database := Database{
		sqlite: FxArgs.SqliteClient,
	}

	models := []any{
		(*SchemaFunSlipper)(nil),
	}

	for _, model := range models {
		_, err := FxArgs.SqliteClient.DB.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(context.Background())
		if err != nil {
			return nil, repoerrors.TranslateSqliteError(err)
		}
	}

	return &database, nil
}

func (database *Database) IncrementCount(ctx context.Context, funslipper domain.FunSlipper) (*domain.FunSlipper, error) {
	schema := NewSchemaFunSplipper(funslipper)
	err := database.sqlite.DB.NewInsert().
		Model(schema).
		On("CONFLICT (user_id, guild_id) DO UPDATE").
		Set("count = count + 1").
		Returning("*").
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainFunSlipper(), nil
}
