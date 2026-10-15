package embeds

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
	CreateOne(ctx context.Context, embed domain.Embed) (*domain.Embed, error)
	UpsertOne(ctx context.Context, embed domain.Embed) (*domain.Embed, error)
	DeleteOne(ctx context.Context, id int) (*domain.Embed, error)
	GetOne(ctx context.Context, id int) (*domain.Embed, error)
	GetAll(ctx context.Context, guildID, authorID string) ([]domain.Embed, error)
}

func New(params Params) (*Database, error) {
	database := Database{
		sqlite: params.SqliteClient,
	}

	models := []any{
		(*SchemaEmbed)(nil),
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

func (database *Database) CreateOne(ctx context.Context, embed domain.Embed) (*domain.Embed, error) {
	schema := NewSchemaEmbeds(embed)

	_, err := database.sqlite.DB.NewInsert().Model(schema).Returning("*").Exec(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainEmbed(), nil
}

func (database *Database) UpsertOne(ctx context.Context, embed domain.Embed) (*domain.Embed, error) {
	schema := NewSchemaEmbeds(embed)

	_, err := database.sqlite.DB.NewInsert().Model(schema).On("CONFLICT (id) DO UPDATE").Set("content = ?", embed.Content).Returning("*").Exec(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainEmbed(), nil
}

func (database *Database) DeleteOne(ctx context.Context, id int) (*domain.Embed, error) {
	var schema SchemaEmbed
	err := database.sqlite.DB.NewDelete().Model(&schema).Where("id = ?", id).Returning("*").Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}
	return schema.NewDomainEmbed(), nil
}

func (database *Database) GetOne(ctx context.Context, id int) (*domain.Embed, error) {
	var schema SchemaEmbed
	err := database.sqlite.DB.NewSelect().Model(&schema).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}
	return schema.NewDomainEmbed(), nil
}

func (database *Database) GetAll(ctx context.Context, guildID string, authorID string) ([]domain.Embed, error) {
	var schemas []SchemaEmbed

	err := database.sqlite.DB.NewSelect().Model(&schemas).Where("guild_id = ? && author_id = ?", guildID, authorID).Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	domain := make([]domain.Embed, 0, len(schemas))
	for _, value := range schemas {
		domain = append(domain, *value.NewDomainEmbed())
	}

	return domain, nil
}
