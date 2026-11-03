package dbtemplates

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type test struct {
	suite.Suite
	database *bun.DB
	storage  Interface
	fixture  *dbfixture.Fixture
}

func (test *test) SetupTest() {
	sqlitedb, err := sql.Open(sqliteshim.ShimName, "file:templates?mode=memory&cache=shared")
	test.NoError(err)
	sqlitedb.SetMaxOpenConns(1)

	test.database = bun.NewDB(sqlitedb, sqlitedialect.New())
	test.database.RegisterModel((*SchemaTemplate)(nil))
	test.database.RegisterModel((*SchemaTemplateContent)(nil))

	_, err = test.database.ExecContext(context.Background(), "PRAGMA foreign_keys = ON")
	test.NoError(err)

	storage, err := New(FxArgs{
		SqliteClient: &sqliteclient.SqliteClient{
			DB: test.database,
		},
	})
	test.NoError(err)
	test.storage = storage

	test.fixture = dbfixture.New(test.database, dbfixture.WithTruncateTables())
}

func (test *test) loadFixture() {
	err := test.fixture.Load(context.Background(), os.DirFS("./../../../../../testdata"), "fixtemplates.yml")
	test.NoError(err)
}

func (test *test) loadRow(id string) any {
	row, err := test.fixture.Row(id)
	test.NoError(err)
	return row
}

func (test *test) parseUTC(t1, t2 *domain.Template) (*domain.Template, *domain.Template) {
	t1.CreatedAt = t1.CreatedAt.UTC().Truncate(time.Microsecond)
	t1.UpdatedAt = t1.UpdatedAt.UTC().Truncate(time.Microsecond)
	t2.CreatedAt = t2.CreatedAt.UTC().Truncate(time.Microsecond)
	t2.UpdatedAt = t2.UpdatedAt.UTC().Truncate(time.Microsecond)
	t1.TemplateContent.CreatedAt = t1.TemplateContent.CreatedAt.UTC().Truncate(time.Microsecond)
	t1.TemplateContent.UpdatedAt = t1.TemplateContent.UpdatedAt.UTC().Truncate(time.Microsecond)
	t2.TemplateContent.CreatedAt = t2.TemplateContent.CreatedAt.UTC().Truncate(time.Microsecond)
	t2.TemplateContent.UpdatedAt = t2.TemplateContent.UpdatedAt.UTC().Truncate(time.Microsecond)

	return t1, t2
}

func (test *test) TearDownTest() {
	test.NoError(test.database.Close())
}

func (test *test) TestCreateOne() {
	now := time.Now().UTC().Truncate(time.Microsecond)
	expect := &domain.Template{
		ID:       1,
		GuildID:  "new",
		AuthorID: "new-user",
		TemplateContent: &domain.TemplateContent{
			Content:     "hello",
			Embeds:      "[]",
			Attachments: "[]",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	actual, err := test.storage.CreateOne(context.Background(), *expect)
	test.NoError(err)
	test.NotNil(actual)

	actual, expect = test.parseUTC(actual, expect)
	test.Equal(expect, actual)

}

func (test *test) TestGetOne() {
	test.loadFixture()
	rowTemplate, rowContent :=
		test.loadRow("SchemaTemplate.one").(*SchemaTemplate),
		test.loadRow("SchemaTemplateContent.one").(*SchemaTemplateContent)
	rowTemplate.TemplateContent = rowContent
	expect := rowTemplate.NewDomainTemplate()

	actual, err := test.storage.GetOne(context.Background(), 1)
	test.NoError(err)
	test.NotNil(actual)

	actual, expect = test.parseUTC(actual, expect)
	test.Equal(expect, actual)
}

func (test *test) TestGetOneNotFound() {
	actual, err := test.storage.GetOne(context.Background(), 5)
	test.Error(err)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
	test.Nil(actual)
}

func (test *test) TestGetAll() {
	test.loadFixture()
	expect, err := test.storage.GetAll(context.Background(), "guild", "author")
	test.NoError(err)
	test.NotNil(expect)
	test.Len(expect, 1)
}

func (test *test) TestGetAllNotFound() {
	expect, err := test.storage.GetAll(context.Background(), "two", "bob")
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
	test.Empty(expect)
}

func (test *test) TestUpdateOne() {
	test.loadFixture()
	rowTemplate := test.loadRow("SchemaTemplate.one").(*SchemaTemplate)
	rowContent := test.loadRow("SchemaTemplateContent.one").(*SchemaTemplateContent)
	rowContent.Content = "new-content"
	rowTemplate.TemplateContent = rowContent
	expect := rowTemplate.NewDomainTemplate()

	actual, err := test.storage.UpdateOne(context.Background(), *rowTemplate.NewDomainTemplate())
	test.NoError(err)
	test.NotNil(actual)

	actual, expect = test.parseUTC(actual, expect)
	test.Equal(expect, actual)

	actual, err = test.storage.GetOne(context.Background(), 1)
	test.NoError(err)
	test.Equal("new-content", actual.TemplateContent.Content)
}

func (test *test) TestUpdateOneNotFound() {
	expect, err := test.storage.UpdateOne(context.Background(), domain.Template{
		ID:       10,
		GuildID:  "bob",
		AuthorID: "bob",
		TemplateContent: &domain.TemplateContent{
			ID: 5,
		},
	})
	test.Error(err)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
	test.Nil(expect)
}

func (test *test) TestDeleteOne() {
	test.loadFixture()

	actual, err := test.storage.DeleteOne(context.Background(), 1)
	test.NoError(err)
	test.NotNil(actual)

	expect, err := test.storage.GetOne(context.Background(), 1)
	test.Error(err)
	test.Nil(expect)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
}

func (test *test) TestDeleteOneNotFound() {
	actual, err := test.storage.DeleteOne(context.Background(), 99)
	test.Error(err)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
	test.Nil(actual)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
