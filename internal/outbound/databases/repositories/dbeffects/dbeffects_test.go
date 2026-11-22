package dbeffects

import (
	"context"
	"database/sql"
	"fmt"
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
	dbName := fmt.Sprintf("file:effect_%d?mode=memory&cache=shared", time.Now().UnixNano())
	sqlitedb, err := sql.Open(sqliteshim.ShimName, dbName)
	test.NoError(err)
	sqlitedb.SetMaxOpenConns(1)

	test.database = bun.NewDB(sqlitedb, sqlitedialect.New())
	test.database.RegisterModel((*SchemaEffect)(nil))

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
	err := test.fixture.Load(context.Background(), os.DirFS("./../../../../../testdata"), "fixeffects.yml")
	test.NoError(err)
}

func (test *test) loadRow(id string) any {
	test.loadFixture()
	row, err := test.fixture.Row(id)
	test.NoError(err)
	return row
}

func (test *test) parseTime(e1, e2 *domain.Effect) (*domain.Effect, *domain.Effect) {
	e1.CreatedAt = e1.CreatedAt.UTC().Truncate(time.Microsecond)
	e1.UpdatedAt = e1.UpdatedAt.UTC().Truncate(time.Microsecond)
	e1.Expiration = e1.Expiration.UTC().Truncate(time.Microsecond)

	e2.CreatedAt = e2.CreatedAt.UTC().Truncate(time.Microsecond)
	e2.UpdatedAt = e2.UpdatedAt.UTC().Truncate(time.Microsecond)
	e2.Expiration = e2.Expiration.UTC().Truncate(time.Microsecond)

	return e1, e2
}

func (test *test) TestGetOne() {
	schema := test.loadRow("SchemaEffect.one").(*SchemaEffect)

	actual, err := test.storage.GetOne(context.Background(), "2123433400718065664", "2123433441100824576")
	test.NoError(err)
	test.NotNil(actual)

	expect := schema.NewDomainEffect()
	actual, expect = test.parseTime(actual, expect)
	test.Equal(actual, expect)
}

func (test *test) TestGetAll() {
	test.loadFixture()
	expect, err := test.storage.GetAll(context.Background(), "2123433400718065664", 3)
	test.NoError(err)
	test.NotNil(expect)
	test.Len(*expect, 3)
}

func (test *test) TestUpdateOne() {
	expect := test.loadRow("SchemaEffect.one").(*SchemaEffect).NewDomainEffect()
	expect.Expiration = time.Now().UTC().Add(time.Hour).Truncate(time.Microsecond)
	expect.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

	actual, err := test.storage.UpdateOne(context.Background(), *expect)
	test.NoError(err)
	test.NotNil(actual)

	actual, expect = test.parseTime(actual, expect)
	test.Equal(expect, actual)

	actual, err = test.storage.GetOne(context.Background(), expect.GuildID, expect.UserID)
	test.NoError(err)
	actual, expect = test.parseTime(actual, expect)
	test.Equal(expect, actual)
}

func (test *test) TestGetOneNotFound() {
	expect, err := test.storage.GetOne(context.Background(), "2123433400718065664", "2123433441100824580")
	test.Error(err)
	test.Nil(expect)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
}

func (test *test) TestCreateOne() {
	now := time.Now().UTC().Truncate(time.Microsecond)
	expiration := time.Now().Add(time.Minute * 30).Truncate(time.Microsecond)

	expect, err := test.storage.CreateOne(context.Background(), domain.Effect{
		UserID:     "2123433441100824576",
		GuildID:    "2123433400718065664",
		Expiration: expiration,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	test.NoError(err)
	test.NotNil(expect)

	actual, err := test.storage.GetOne(context.Background(), "2123433400718065664", "2123433441100824576")
	test.NoError(err)
	test.NotNil(actual)

	actual, expect = test.parseTime(actual, expect)

	test.Equal(expect, actual)
}

func (test *test) TestCreateOneConflict() {
	test.loadFixture()

	expect, err := test.storage.CreateOne(context.Background(), domain.Effect{
		UserID:     "2123433441100824576",
		GuildID:    "2123433400718065664",
		Expiration: time.Now().Add(time.Minute * 30),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})
	test.Error(err)
	test.Nil(expect)

	test.True(core.IsCode(err, repoerrors.CodeSqlConflict))
}

func (test *test) TestDeleteOne() {
	test.loadFixture()

	expect, err := test.storage.DeleteOne(context.Background(), "2123433441100824576")
	test.NoError(err)
	test.NotNil(expect)

	actual, err := test.storage.GetOne(context.Background(), "2123433400718065664", "2123433441100824576")
	test.Error(err)
	test.Nil(actual)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
}

func (test *test) TestDeleteOneNotFound() {
	expect, err := test.storage.DeleteOne(context.Background(), "2123433441100824580")
	test.Error(err)
	test.Nil(expect)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
