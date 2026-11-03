package dbtemplates

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

//go:generate mockgen -source=templates.go -destination=mocks/mock_templates.go -package=mock_templates
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
	CreateOne(ctx context.Context, template domain.Template) (*domain.Template, error)
	GetOne(ctx context.Context, id int) (*domain.Template, error)
	GetAll(ctx context.Context, guildID, authorID string) ([]domain.Template, error)
	UpdateOne(ctx context.Context, template domain.Template) (*domain.Template, error)
	DeleteOne(ctx context.Context, id int) (*domain.Template, error)
}

func New(FxArgs FxArgs) (*Database, error) {
	database := Database{
		sqlite: FxArgs.SqliteClient,
	}

	models := []any{
		(*SchemaTemplate)(nil),
		(*SchemaTemplateContent)(nil),
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

func (database *Database) CreateOne(ctx context.Context, template domain.Template) (*domain.Template, error) {
	if template.TemplateContent == nil {
		return nil, repoerrors.NewErrorSqlUnknown(errors.New("template content is nil"))
	}
	templateSchema := NewSchemaTemplate(template)
	contentSchema := NewSchemaTemplateContent(*template.TemplateContent)
	err := database.sqlite.DB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {

		err := tx.NewInsert().
			Model(&templateSchema).
			Returning("id").
			Scan(ctx, &templateSchema.ID)
		if err != nil {
			return err
		}

		contentSchema.TemplateID = templateSchema.ID
		err = tx.NewInsert().
			Model(&contentSchema).
			Returning("*").
			Scan(ctx, &contentSchema)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	template.ID = templateSchema.ID
	template.TemplateContent.ID = contentSchema.ID
	template.TemplateContent.TemplateID = contentSchema.TemplateID
	return &template, nil
}

func (database *Database) GetOne(ctx context.Context, id int) (*domain.Template, error) {
	var schema SchemaTemplate
	err := database.sqlite.DB.NewSelect().
		Model(&schema).
		Relation("TemplateContent").
		Where("t.id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return schema.NewDomainTemplate(), nil
}

func (database *Database) GetAll(ctx context.Context, guildID string, authorID string) ([]domain.Template, error) {
	var schemas []SchemaTemplate

	err := database.sqlite.DB.NewSelect().
		Model(&schemas).
		Where("guild_id = ? AND author_id = ?", guildID, authorID).
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	if len(schemas) == 0 {
		return nil, repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)
	}

	domain := make([]domain.Template, 0, len(schemas))
	for _, value := range schemas {
		domain = append(domain, *value.NewDomainTemplate())
	}

	return domain, nil
}

func (database *Database) UpdateOne(ctx context.Context, template domain.Template) (*domain.Template, error) {
	if template.TemplateContent == nil {
		return nil, repoerrors.NewErrorSqlUnknown(errors.New("template content is nil"))
	}

	schema := NewSchemaTemplate(template)
	contentSchema := NewSchemaTemplateContent(*template.TemplateContent)
	contentSchema.TemplateID = template.ID
	err := database.sqlite.DB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		result, err := tx.NewUpdate().
			Model(&schema).
			Column("guild_id", "author_id", "updated_at").
			Where("id = ?", schema.ID).
			Exec(ctx)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		result, err = tx.NewUpdate().
			Model(&contentSchema).
			Column("content", "attachments", "embeds", "updated_at").
			Where("template_id = ?", schema.ID).
			Exec(ctx)
		if err != nil {
			return err
		}

		rowsAffected, err = result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	return &template, nil
}

func (database *Database) DeleteOne(ctx context.Context, id int) (*domain.Template, error) {
	var template SchemaTemplate
	var content SchemaTemplateContent

	err := database.sqlite.DB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		result, err := tx.NewDelete().
			Model((*SchemaTemplate)(nil)).
			Where("id = ?", id).
			Returning("*").
			Exec(ctx, &template)
		if err != nil {
			return repoerrors.TranslateSqliteError(err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return repoerrors.TranslateSqliteError(err)
		}
		if rowsAffected == 0 {
			return repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)
		}

		result, err = tx.NewDelete().
			Model((*SchemaTemplateContent)(nil)).
			Where("template_id = ?", id).
			Returning("*").
			Exec(ctx, &content)
		if err != nil {
			return repoerrors.TranslateSqliteError(err)
		}

		rowsAffected, err = result.RowsAffected()
		if err != nil {
			return repoerrors.TranslateSqliteError(err)
		}
		if rowsAffected == 0 {
			return repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)
		}

		return nil
	})

	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	template.TemplateContent = &content
	return template.NewDomainTemplate(), nil
}
