package server

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type Database struct {
	sqlite *sqliteclient.SqliteClient
}

type Interface interface {
	GetServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error)
	CreateServerWithConfig(ctx context.Context, server domain.Server) error
}

type FxArgs struct {
	fx.In

	SqliteClient *sqliteclient.SqliteClient
}

func New(FxArgs FxArgs) (*Database, error) {
	database := Database{
		sqlite: FxArgs.SqliteClient,
	}

	models := []any{
		(*SchemaServer)(nil),
		(*SchemaConfig)(nil),
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

func (database *Database) CreateServerWithConfig(
	ctx context.Context,
	server domain.Server,
) error {
	if server.Config == nil {
		return errors.New("server config is required")
	}

	err := database.sqlite.DB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
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
	if err != nil {
		return repoerrors.TranslateSqliteError(err)
	}

	return nil
}

func (database *Database) GetServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error) {
	schema := new(SchemaServer)

	err := database.sqlite.DB.NewSelect().
		Model(schema).
		Relation("Config").
		Where("server.guild_id = ?", guildID).
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
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
