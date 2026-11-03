package dbtomatoes

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/core"
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
	sqlitedb, err := sql.Open(sqliteshim.ShimName, "file:tomatoes?mode=memory&cache=shared")
	test.NoError(err)
	sqlitedb.SetMaxOpenConns(1)

	test.database = bun.NewDB(sqlitedb, sqlitedialect.New())
	test.database.RegisterModel((*SchemaTomatoes)(nil))

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
	err := test.fixture.Load(context.Background(), os.DirFS("./../../../../../testdata"), "fixtomatoes.yml")
	test.NoError(err)
}

func (test *test) loadRow(id string) any {
	row, err := test.fixture.Row(id)
	test.NoError(err)
	return row
}

func (test *test) TestIncrement() {
	test.loadFixture()
	actual, err := test.storage.Increment(context.Background(), "one", "user_one", 5)
	test.NoError(err)
	test.Equal(15, actual)

	actual, err = test.storage.Increment(context.Background(), "one", "user_one", 5)
	test.NoError(err)
	test.Equal(20, actual)
}

func (test *test) TestDecrement() {
	test.loadFixture()
	actual, err := test.storage.Decrement(context.Background(), "one", "user_one", 5)
	test.NoError(err)
	test.Equal(actual, 5)

	actual, err = test.storage.Decrement(context.Background(), "one", "user_one", 5)
	test.NoError(err)
	test.Equal(actual, 0)
}

func (test *test) TestGetInventory() {
	test.loadFixture()
	expect := test.loadRow("SchemaTomatoes.one").(*SchemaTomatoes)
	actual, err := test.storage.GetInventory(test.T().Context(), "one", "user_one")
	test.NoError(err)
	test.Equal(actual, expect.Inventory)
}

func (test *test) TestGetInventoryNotFound() {
	actual, err := test.storage.GetInventory(test.T().Context(), "one", "non_existent_user")
	test.Error(err)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
	test.Zero(actual)
}

func (test *test) TearDownTest() {
	test.NoError(test.database.Close())
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
