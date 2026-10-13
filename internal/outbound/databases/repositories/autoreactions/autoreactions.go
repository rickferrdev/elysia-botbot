package autoreactions

import (
	"context"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqlite"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Database struct {
		sqlite *sqlite.SqliteClient
	}

	Params struct {
		fx.In
		SqliteClient *sqlite.SqliteClient
	}
)

type Interface interface {
	CreateOne(ctx context.Context, autoreactions domain.AutoReactions) (*domain.AutoReactions, error)
	DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error)
	GetOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error)
	UpdateOne(ctx context.Context, guildID, channelID, emoji string) (*domain.AutoReactions, error)
	GetAll(ctx context.Context, guildID string) ([]domain.AutoReactions, error)
}

func New(params Params) (*Database, error) {
	database := Database{
		sqlite: params.SqliteClient,
	}

	models := []any{
		(*SchemaAutoReactions)(nil),
	}

	for _, model := range models {
		_, err := params.SqliteClient.DB.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(context.Background())
		if err != nil {
			return nil, repoerrors.TranslateSqliteError(err)
		}
	}

	return &database, nil
}

func (database *Database) CreateOne(ctx context.Context, autoreactions domain.AutoReactions) (*domain.AutoReactions, error) {
	schema := NewSchemaAutoReactions(autoreactions)

	_, err := database.sqlite.DB.NewInsert().Model(schema).Returning("*").Exec(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainAutoReactions(), nil
}

func (database *Database) GetOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error) {
	var schema SchemaAutoReactions
	err := database.sqlite.DB.NewSelect().Model(&schema).Where("channel_id = ? AND guild_id = ?", channelID, guildID).Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}
	return schema.NewDomainAutoReactions(), nil
}

func (database *Database) GetAll(ctx context.Context, guildID string) ([]domain.AutoReactions, error) {
	var schemas []SchemaAutoReactions

	err := database.sqlite.DB.NewSelect().Model(&schemas).Where("guild_id = ?", guildID).Scan(ctx)
	if err != nil {
		return []domain.AutoReactions{}, repoerrors.TranslateSqliteError(err)
	}

	domain := make([]domain.AutoReactions, 0, len(schemas))
	for _, value := range schemas {
		domain = append(domain, *value.NewDomainAutoReactions())
	}

	return domain, nil
}

func (database *Database) UpdateOne(ctx context.Context, guildID, channelID, emoji string) (*domain.AutoReactions, error) {
	var schema SchemaAutoReactions
	err := database.sqlite.DB.NewUpdate().Model(&schema).Set("emoji = ?", emoji).Where("channel_id = ? AND guild_id = ?", channelID, guildID).Returning("*").Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}
	return schema.NewDomainAutoReactions(), nil
}

func (database *Database) DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReactions, error) {
	var schema SchemaAutoReactions
	err := database.sqlite.DB.NewDelete().Model(&schema).Where("channel_id = ? AND guild_id = ?", channelID, guildID).Returning("*").Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}
	return schema.NewDomainAutoReactions(), nil
}
