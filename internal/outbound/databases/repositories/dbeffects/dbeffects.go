package dbeffects

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type SchemaEffect struct {
	bun.BaseModel `bun:"table:effects"`

	ID         int               `bun:"id,pk,autoincrement"`
	Type       domain.EffectType `bun:"type,notnull"`
	UserID     string            `bun:"user_id,notnull,unique"`
	GuildID    string            `bun:"guild_id,notnull"`
	Expiration time.Time         `bun:"expiration,notnull"`
	CreatedAt  time.Time         `bun:"created_at"`
	UpdatedAt  time.Time         `bun:"updated_at"`
}

func NewSchemaEffect(effect domain.Effect) *SchemaEffect {
	schema := SchemaEffect{
		ID:         effect.ID,
		Type:       effect.Type,
		UserID:     effect.UserID,
		GuildID:    effect.GuildID,
		Expiration: effect.Expiration,
		CreatedAt:  effect.CreatedAt,
		UpdatedAt:  effect.UpdatedAt,
	}

	return &schema
}

func (schema *SchemaEffect) NewDomainEffect() *domain.Effect {
	domain := domain.Effect{
		ID:         schema.ID,
		Type:       schema.Type,
		UserID:     schema.UserID,
		GuildID:    schema.GuildID,
		Expiration: schema.Expiration,
		CreatedAt:  schema.CreatedAt,
		UpdatedAt:  schema.UpdatedAt,
	}

	return &domain
}

//go:generate go run go.uber.org/mock/mockgen -source=dbeffects.go -destination=mocks/mock_effects.go -package=mock_effects
type (
	Database struct {
		sqlite *sqliteclient.SqliteClient
	}

	FxArgs struct {
		fx.In
		SqliteClient *sqliteclient.SqliteClient
	}

	Interface interface {
		GetOne(ctx context.Context, userID string) (*domain.Effect, error)
		CreateOne(ctx context.Context, effect domain.Effect) (*domain.Effect, error)
		UpdateOne(ctx context.Context, effect domain.Effect) (*domain.Effect, error)
		DeleteOne(ctx context.Context, userID string) (*domain.Effect, error)
		GetAll(ctx context.Context, guildID string, limit int) (*[]domain.Effect, error)
	}
)

func New(args FxArgs) (*Database, error) {
	database := Database{
		sqlite: args.SqliteClient,
	}

	models := []any{
		(*SchemaEffect)(nil),
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

func (database *Database) GetOne(ctx context.Context, userID string) (*domain.Effect, error) {
	var schema SchemaEffect
	err := database.sqlite.DB.NewSelect().
		Model(&schema).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)
		}
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainEffect(), nil
}

func (database *Database) GetAll(ctx context.Context, guildID string, limit int) (*[]domain.Effect, error) {
	if limit <= 0 {
		limit = 100
	}

	schemas := make([]SchemaEffect, 0)
	if err := database.sqlite.DB.NewSelect().
		Model(&schemas).
		Where("guild_id = ?", guildID).
		Limit(int64(limit)).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)
		}
		return nil, repoerrors.TranslateSqliteError(err)
	}

	if len(schemas) == 0 {
		return nil, repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)
	}

	domain := make([]domain.Effect, 0, len(schemas))
	for _, val := range schemas {
		domain = append(domain, *val.NewDomainEffect())
	}

	return &domain, nil
}

func (database *Database) CreateOne(ctx context.Context, effect domain.Effect) (*domain.Effect, error) {
	schema := NewSchemaEffect(effect)

	_, err := database.sqlite.DB.NewInsert().
		Model(schema).
		Exec(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainEffect(), nil
}

func (database *Database) UpdateOne(ctx context.Context, effect domain.Effect) (*domain.Effect, error) {
	schema := NewSchemaEffect(effect)

	if err := database.sqlite.DB.NewUpdate().
		Model(schema).
		Returning("*").
		Set("type = ?", schema.Type).
		Set("guild_id = ?", schema.GuildID).
		Set("expiration = ?", schema.Expiration).
		Set("updated_at = ?", schema.UpdatedAt).
		Where("user_id = ?", schema.UserID).
		Scan(ctx, schema); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)
		}
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainEffect(), nil
}

func (database *Database) DeleteOne(ctx context.Context, userID string) (*domain.Effect, error) {
	var schema SchemaEffect

	err := database.sqlite.DB.NewDelete().
		Model(&schema).
		Where("user_id = ?", userID).
		Returning("*").
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainEffect(), nil
}
