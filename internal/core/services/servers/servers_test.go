package servers

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	mock_servers "github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbservers/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type test struct {
	suite.Suite
	storage *mock_servers.MockInterface
	service *Service
}

func (test *test) SetupTest() {
	test.storage = mock_servers.NewMockInterface(gomock.NewController(test.T()))
	test.service = New(FxArgs{ServerStorage: test.storage})
}

func (test *test) TestCreateServerWithConfig() {
	ctx := context.Background()
	server := domain.Server{
		GuildID: "guild-id",
		Config:  &domain.Config{GuildID: "guild-id", Prefix: "!"},
	}

	test.storage.EXPECT().CreateServerWithConfig(ctx, server).Return(&server, nil)

	err := test.service.CreateServerWithConfig(ctx, server)
	test.NoError(err)
}

func (test *test) TestCreateServerWithConfigConflict() {
	ctx := context.Background()
	server := domain.Server{GuildID: "guild-id"}
	conflict := repoerrors.NewErrorSqlConflict(errors.New("conflict"))

	test.storage.EXPECT().CreateServerWithConfig(ctx, server).Return(nil, conflict)

	err := test.service.CreateServerWithConfig(ctx, server)
	test.ErrorIs(err, ErrServerConfigurationAlreadyExists)
}

func (test *test) TestObtainServerWithConfig() {
	ctx := context.Background()
	expect := &domain.Server{
		GuildID: "guild-id",
		Config:  &domain.Config{GuildID: "guild-id", Prefix: "!"},
	}

	test.storage.EXPECT().GetServerWithConfig(ctx, "guild-id").Return(expect, nil)

	actual, err := test.service.ObtainServerWithConfig(ctx, "guild-id")
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestObtainServerWithConfigNotFound() {
	ctx := context.Background()
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.storage.EXPECT().GetServerWithConfig(ctx, "guild-id").Return(nil, notFound)

	actual, err := test.service.ObtainServerWithConfig(ctx, "guild-id")
	test.Nil(actual)
	test.ErrorIs(err, ErrServerConfigurationNotFound)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
