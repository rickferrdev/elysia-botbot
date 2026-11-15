package dbkwaii

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
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
	sqlitedb, err := sql.Open(sqliteshim.ShimName, "file:kwaii?mode=memory&cache=shared")
	test.NoError(err)
	sqlitedb.SetMaxOpenConns(1)

	test.database = bun.NewDB(sqlitedb, sqlitedialect.New())
	test.database.RegisterModel((*SchemaKwaii)(nil))

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
	err := test.fixture.Load(context.Background(), os.DirFS("./../../../../../../testdata"), "fixkwaii.yml")
	test.NoError(err)
}

func (test *test) loadRow(id string) any {
	row, err := test.fixture.Row(id)
	test.NoError(err)
	return row
}

func (test *test) TestCreateOne() {
	expect, err := test.storage.CreateOne(context.Background(), "userID")
	test.NoError(err)
	test.NotNil(expect)
}

func (test *test) TestDeleteOne() {
	test.loadFixture()
	expect, err := test.storage.DeleteOne(context.Background(), "123")
	test.NoError(err)
	test.NotNil(expect)
}

func (test *test) TestGetOne() {
	test.loadFixture()
	expect := test.loadRow("SchemaKwaii.one").(*SchemaKwaii)

	actual, err := test.storage.GetOne(context.Background(), "123")
	test.NoError(err)
	test.NotNil(actual)

	test.Equal(expect.NewDomainKwaii(), actual)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
