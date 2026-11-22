package dbautoreactions

import (
	"context"
	"database/sql"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type SchemaAutoReactions struct {
	bun.BaseModel `bun:"table:autoreactions,alias:autoreaction"`

	ID        int                     `bun:"id,pk,autoincrement"`
	GuildID   string                  `bun:"guild_id,notnull"`
	ChannelID string                  `bun:"channel_id,unique,notnull"`
	Type      domain.AutoReactionType `bun:"type,notnull"`
	Emoji     string                  `bun:"emoji,notnull"`
}

func NewSchemaAutoReactions(autoreactions domain.AutoReaction) *SchemaAutoReactions {
	schema := SchemaAutoReactions{
		ID:        autoreactions.ID,
		GuildID:   autoreactions.GuildID,
		ChannelID: autoreactions.ChannelID,
		Type:      autoreactions.Type,
		Emoji:     autoreactions.Emoji,
	}

	return &schema
}

func (schema *SchemaAutoReactions) NewDomainAutoReactions() *domain.AutoReaction {
	domain := domain.AutoReaction{
		ID:        schema.ID,
		GuildID:   schema.GuildID,
		ChannelID: schema.ChannelID,
		Type:      schema.Type,
		Emoji:     schema.Emoji,
	}

	return &domain
}

//go:generate mockgen -source=autoreactions.go -destination=mocks/mock_autoreactions.go -package=mock_autoreactions
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
	GetAll(ctx context.Context, guildID string) ([]domain.AutoReaction, error)
	GetOne(ctx context.Context, guildID, channelID string) (*domain.AutoReaction, error)
	CreateOne(ctx context.Context, autoreactions domain.AutoReaction) (*domain.AutoReaction, error)
	UpdateOne(ctx context.Context, autoreactions domain.AutoReaction) (*domain.AutoReaction, error)
	DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReaction, error)
}

func New(args FxArgs) (*Database, error) {
	database := Database{
		sqlite: args.SqliteClient,
	}

	models := []any{
		(*SchemaAutoReactions)(nil),
	}

	for _, model := range models {
		_, err := args.SqliteClient.DB.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(context.Background())
		if err != nil {
			return nil, repoerrors.TranslateSqliteError(err)
		}
	}

	return &database, nil
}

func (database *Database) CreateOne(ctx context.Context, autoreactions domain.AutoReaction) (*domain.AutoReaction, error) {
	schema := NewSchemaAutoReactions(autoreactions)

	_, err := database.sqlite.DB.NewInsert().
		Model(schema).
		On("CONFLICT (channel_id) DO UPDATE").
		Set("emoji = EXCLUDED.emoji").
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainAutoReactions(), nil
}

func (database *Database) GetOne(ctx context.Context, guildID, channelID string) (*domain.AutoReaction, error) {
	var schema SchemaAutoReactions

	err := database.sqlite.DB.NewSelect().
		Model(&schema).
		Where("channel_id = ? AND guild_id = ?", channelID, guildID).
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainAutoReactions(), nil
}

func (database *Database) GetAll(ctx context.Context, guildID string) ([]domain.AutoReaction, error) {
	var schemas []SchemaAutoReactions

	err := database.sqlite.DB.
		NewSelect().
		Model(&schemas).
		Where("guild_id = ?", guildID).
		Scan(ctx)
	if err != nil {
		return []domain.AutoReaction{}, repoerrors.TranslateSqliteError(err)
	}

	if len(schemas) == 0 {
		return []domain.AutoReaction{}, repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)
	}

	domain := make([]domain.AutoReaction, 0, len(schemas))
	for _, value := range schemas {
		domain = append(domain, *value.NewDomainAutoReactions())
	}

	return domain, nil
}

func (database *Database) UpdateOne(ctx context.Context, autoreactions domain.AutoReaction) (*domain.AutoReaction, error) {
	schema := NewSchemaAutoReactions(autoreactions)

	err := database.sqlite.DB.NewUpdate().
		Model(schema).
		Where("guild_id = ? AND channel_id = ?", schema.GuildID, schema.ChannelID).
		Set("emoji = ?", schema.Emoji).
		Set("type = ?", schema.Type).
		Returning("*").
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainAutoReactions(), nil
}

func (database *Database) DeleteOne(ctx context.Context, guildID, channelID string) (*domain.AutoReaction, error) {
	var schema SchemaAutoReactions

	err := database.sqlite.DB.NewDelete().
		Model(&schema).
		Where("channel_id = ? AND guild_id = ?", channelID, guildID).
		Returning("*").
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainAutoReactions(), nil
}
