package dbservers

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
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
	sqlitedb, err := sql.Open(sqliteshim.ShimName, "file:server?mode=memory&cache=shared")
	test.NoError(err)
	sqlitedb.SetMaxOpenConns(1)

	test.database = bun.NewDB(sqlitedb, sqlitedialect.New())
	test.database.RegisterModel((*SchemaServer)(nil), (*SchemaConfig)(nil))

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
	err := test.fixture.Load(
		context.Background(),
		os.DirFS("./../../../../../testdata"),
		"fixservers.yml",
	)
	test.Require().NoError(err)
}

func (test *test) TearDownTest() {
	test.NoError(test.database.Close())
}

func (test *test) loadRow(id string) any {
	row, err := test.fixture.Row(id)
	test.NoError(err)
	return row
}

func (test *test) parseServerUTC(t1, t2 *domain.Server) (*domain.Server, *domain.Server) {
	t1.CreatedAt = t1.CreatedAt.UTC().Truncate(time.Microsecond)
	t1.UpdatedAt = t1.UpdatedAt.UTC().Truncate(time.Microsecond)
	t2.CreatedAt = t2.CreatedAt.UTC().Truncate(time.Microsecond)
	t2.UpdatedAt = t2.UpdatedAt.UTC().Truncate(time.Microsecond)
	t1.Config.CreatedAt = t1.Config.CreatedAt.UTC().Truncate(time.Microsecond)
	t1.Config.UpdatedAt = t1.Config.UpdatedAt.UTC().Truncate(time.Microsecond)
	t2.Config.CreatedAt = t2.Config.CreatedAt.UTC().Truncate(time.Microsecond)
	t2.Config.UpdatedAt = t2.Config.UpdatedAt.UTC().Truncate(time.Microsecond)

	return t1, t2
}

func (test *test) TestCreateServerWithConfig() {
	guildID := uuid.NewString()
	now := time.Now().UTC().Truncate(time.Microsecond)
	expect := &domain.Server{
		GuildID:   guildID,
		CreatedAt: now,
		UpdatedAt: now,
		Config: &domain.Config{
			Prefix:    "e!",
			GuildID:   guildID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	actual, err := test.storage.CreateServerWithConfig(context.Background(), *expect)
	test.NoError(err)
	test.NotNil(actual)

	actual, expect = test.parseServerUTC(actual, expect)
	test.Equal(expect, actual)
}

func (test *test) TestCreateServerWithInvalidConfig() {
	actual := &domain.Server{
		GuildID: "one",
		Config: &domain.Config{
			Prefix:  "e!",
			GuildID: "two",
		},
	}

	data, expect := test.storage.CreateServerWithConfig(context.Background(), *actual)
	test.Nil(data)
	test.True(core.IsCode(expect, repoerrors.CodeSqlInvalidData))
}

func (test *test) TestGetServerWithConfig() {
	test.loadFixture()
	serverRow := test.loadRow("SchemaServer.one").(*SchemaServer)
	configRow := test.loadRow("SchemaConfig.one").(*SchemaConfig)
	serverRow.Config = configRow
	expect := serverRow.NewDomainServer()

	actual, err := test.storage.GetServerWithConfig(context.Background(), "one")
	test.NoError(err)
	test.NotNil(actual)

	actual, expect = test.parseServerUTC(actual, expect)
	test.Equal(expect, actual)
}

func (test *test) TestGetServerWithConfigNotFound() {
	data, err := test.storage.GetServerWithConfig(context.Background(), "one")

	test.True(core.IsCode(err, repoerrors.CodeSqlNoRows))
	test.Nil(data)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
