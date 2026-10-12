package server

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqlite"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type Database struct {
	sqlite *sqlite.SqliteClient
}

type Interface interface {
	GetServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error)
	CreateServerWithConfig(ctx context.Context, server domain.Server) error
}

type Params struct {
	fx.In
	SqliteClient *sqlite.SqliteClient
}

func New(params Params) (*Database, error) {
	database := Database{
		sqlite: params.SqliteClient,
	}

	models := []any{
		(*SchemaServer)(nil),
		(*SchemaConfig)(nil),
	}

	for _, model := range models {
		_, err := params.SqliteClient.DB.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return &database, nil
}

func (database *Database) CreateServerWithConfig(
	ctx context.Context,
	server domain.Server,
) error {
	if server.Config == nil {
		return errors.New("server config is required")
	}

	return database.sqlite.DB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		serverSchema := NewSchemaServer(&server)
		if _, err := tx.NewInsert().Model(serverSchema).On("CONFLICT (guild_id) DO NOTHING").Exec(ctx); err != nil {
			return err
		}

		configSchema := NewSchemaConfig(server.Config)
		if _, err := tx.NewInsert().Model(configSchema).On("CONFLICT (guild_id) DO UPDATE").Set("prefix = EXCLUDED.prefix").Exec(ctx); err != nil {
			return err
		}
		return nil
	})
}

func (database *Database) GetServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error) {
	schema := new(SchemaServer)

	err := database.sqlite.DB.NewSelect().
		Model(schema).
		Relation("Config").
		Where("server.guild_id = ?", guildID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := &domain.Server{
		ID:      schema.ID,
		GuildID: schema.GuildID,
	}

	if schema.Config != nil {
		result.Config = &domain.Config{
			ID:      schema.Config.ID,
			GuildID: schema.Config.GuildID,
			Prefix:  schema.Config.Prefix,
		}
	}

	return result, nil
}
