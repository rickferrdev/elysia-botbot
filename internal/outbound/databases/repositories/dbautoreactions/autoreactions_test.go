package dbautoreactions

import (
	"context"
	"database/sql"
	"os"
	"testing"

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
	sqlitedb, err := sql.Open(sqliteshim.ShimName, "file:autoreactions?mode=memory&cache=shared")
	test.NoError(err)
	sqlitedb.SetMaxOpenConns(1)

	test.database = bun.NewDB(sqlitedb, sqlitedialect.New())
	test.database.RegisterModel((*SchemaAutoReactions)(nil))

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
	err := test.fixture.Load(context.Background(), os.DirFS("./../../../../../testdata"), "fixautoreactions.yml")
	test.NoError(err)
}

func (test *test) loadRow(id string) any {
	row, err := test.fixture.Row(id)
	test.NoError(err)
	return row
}

func (test *test) TestCreateOne() {
	expect := domain.AutoReactions{
		ID:        3,
		GuildID:   "one",
		ChannelID: "channel_three",
		Type:      domain.AutoReactionTypeMessage,
		Emoji:     "🤖",
	}

	actual, err := test.storage.CreateOne(context.Background(), expect)
	test.NoError(err)
	test.NotNil(actual)
	test.Equal(actual, &expect)
}

func (test *test) TestCreateUpsert() {
	expect := domain.AutoReactions{
		ID:        3,
		GuildID:   "one",
		ChannelID: "channel_two",
		Type:      domain.AutoReactionTypeMessage,
		Emoji:     "🩴",
	}

	actual, err := test.storage.CreateOne(context.Background(), expect)
	test.NoError(err)
	test.NotNil(actual)
	test.Equal(actual, &expect)
}

func (test *test) TestGetOne() {
	test.loadFixture()
	expect := test.loadRow("SchemaAutoReactions.one").(*SchemaAutoReactions).NewDomainAutoReactions()
	actual, err := test.storage.GetOne(context.Background(), "one", "channel_one")
	test.NoError(err)
	test.NotNil(actual)
	test.Equal(expect, actual)
}

func (test *test) TestGetOneNotFound() {
	expect, err := test.storage.GetOne(context.Background(), "one", "channel_one")
	test.Error(err)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
	test.Nil(expect)
}

func (test *test) TestGetAll() {
	test.loadFixture()
	autoReactionsOne, autoReactionsTwo := test.loadRow("SchemaAutoReactions.one").(*SchemaAutoReactions), test.loadRow("SchemaAutoReactions.two").(*SchemaAutoReactions)
	expect := []domain.AutoReactions{
		*autoReactionsOne.NewDomainAutoReactions(),
		*autoReactionsTwo.NewDomainAutoReactions(),
	}

	actual, err := test.storage.GetAll(context.Background(), "one")
	test.NoError(err)
	test.NotNil(actual)
	test.Equal(expect, actual)
}

func (test *test) TestGetAllNotFound() {
	actual, err := test.storage.GetAll(context.Background(), "one")
	test.Error(err)
	test.Empty(actual)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
}

func (test *test) TestUpdateOne() {
	test.loadFixture()
	expect := &domain.AutoReactions{
		ID:        1,
		GuildID:   "one",
		ChannelID: "channel_one",
		Type:      domain.AutoReactionTypeAttachment,
		Emoji:     "🤖",
	}

	actual, err := test.storage.UpdateOne(context.Background(), *expect)
	test.NoError(err)
	test.NotNil(actual)

	test.Equal(expect, actual)
}

func (test *test) TestUpdateNotFound() {
	expect := &domain.AutoReactions{
		ID:        10,
		GuildID:   "one",
		ChannelID: "channel_one",
		Type:      domain.AutoReactionTypeAttachment,
		Emoji:     "🤖",
	}

	actual, err := test.storage.UpdateOne(context.Background(), *expect)
	test.Error(err)
	test.Empty(actual)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
}

func (test *test) TestDeleteOne() {
	test.loadFixture()
	expect := test.loadRow("SchemaAutoReactions.one").(*SchemaAutoReactions).NewDomainAutoReactions()
	actual, err := test.storage.DeleteOne(context.Background(), "one", "channel_one")
	test.NoError(err)
	test.NotNil(actual)
	test.Equal(expect, actual)
}

func (test *test) TestDeleteOneNotFound() {
	expect, err := test.storage.DeleteOne(context.Background(), "one", "channel_one")
	test.Error(err)
	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
	test.Nil(expect)
}

func (test *test) TearDownTest() {
	test.NoError(test.database.Close())
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
