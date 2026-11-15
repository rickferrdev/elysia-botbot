package dbkwaii

import (
	"context"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type SchemaKwaii struct {
	bun.BaseModel `bun:"table:kwaii"`
	ID            int    `bun:"id,pk,autoincrement"`
	UserID        string `bun:"user_id,notnull,unique"`
}

func NewSchemaKwaii(kwaii domain.Kwaii) *SchemaKwaii {
	schema := SchemaKwaii{
		ID:     kwaii.ID,
		UserID: kwaii.UserID,
	}

	return &schema
}

func (schema *SchemaKwaii) NewDomainKwaii() *domain.Kwaii {
	domain := domain.Kwaii{
		ID:     schema.ID,
		UserID: schema.UserID,
	}

	return &domain
}

type (
	Database struct {
		sqlite *sqliteclient.SqliteClient
	}

	FxArgs struct {
		fx.In
		SqliteClient *sqliteclient.SqliteClient
	}

	Interface interface {
		GetOne(ctx context.Context, userID string) (*domain.Kwaii, error)
		CreateOne(ctx context.Context, userID string) (*domain.Kwaii, error)
		DeleteOne(ctx context.Context, userID string) (*domain.Kwaii, error)
	}
)

func New(args FxArgs) (*Database, error) {
	database := Database{
		sqlite: args.SqliteClient,
	}

	models := []any{
		(*SchemaKwaii)(nil),
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

func (database *Database) GetOne(ctx context.Context, userID string) (*domain.Kwaii, error) {
	var schema SchemaKwaii
	err := database.sqlite.DB.NewSelect().
		Model(&schema).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainKwaii(), nil
}

func (database *Database) CreateOne(ctx context.Context, userID string) (*domain.Kwaii, error) {
	schema := NewSchemaKwaii(domain.Kwaii{
		UserID: userID,
	})

	_, err := database.sqlite.DB.NewInsert().
		Model(schema).
		Exec(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainKwaii(), nil
}

func (database *Database) DeleteOne(ctx context.Context, userID string) (*domain.Kwaii, error) {
	var schema SchemaKwaii
	err := database.sqlite.DB.NewDelete().
		Model(&schema).
		Where("user_id = ?", userID).
		Returning("*").
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainKwaii(), nil
}
