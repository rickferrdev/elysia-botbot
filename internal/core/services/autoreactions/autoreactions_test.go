package autoreactions

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	mock_autoreactions "github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbautoreactions/mocks"
	mock_cache "github.com/rickferrdev/elysia-botbot/internal/platform/cache/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type test struct {
	suite.Suite
	storage *mock_autoreactions.MockInterface
	cache   *mock_cache.MockInterface
	service *Service
}

func (test *test) SetupTest() {
	controller := gomock.NewController(test.T())
	test.storage = mock_autoreactions.NewMockInterface(controller)
	test.cache = mock_cache.NewMockInterface(controller)
	test.service = New(FxArgs{
		Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		Cache:           test.cache,
		AutoReactionsDB: test.storage,
	})
}

func (test *test) TestCreateOne() {
	ctx := context.Background()
	expect := &domain.AutoReaction{
		ID:        1,
		GuildID:   "guild-id",
		ChannelID: "channel-id",
		Type:      domain.AutoReactionTypeMessage,
		Emoji:     "🤖",
	}

	test.storage.EXPECT().CreateOne(ctx, *expect).Return(expect, nil)
	test.cache.EXPECT().Del(ctx, "autoreactions:guild-id:channel-id", "autoreactions:guild:guild-id").Return(nil)

	actual, err := test.service.CreateOne(ctx, *expect)
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestCreateOneConflict() {
	ctx := context.Background()
	autoreaction := domain.AutoReaction{GuildID: "guild-id", ChannelID: "channel-id"}
	conflict := repoerrors.NewErrorSqlConflict(errors.New("conflict"))

	test.storage.EXPECT().CreateOne(ctx, autoreaction).Return(nil, conflict)

	actual, err := test.service.CreateOne(ctx, autoreaction)
	test.Nil(actual)
	test.ErrorIs(err, ErrAutoReactionAlreadyExists)
}

func (test *test) TestObtainOne() {
	ctx := context.Background()
	expect := &domain.AutoReaction{
		ID:        1,
		GuildID:   "guild-id",
		ChannelID: "channel-id",
		Type:      domain.AutoReactionTypeMessage,
		Emoji:     "🤖",
	}

	test.cache.EXPECT().GetWithJson(ctx, "autoreactions:guild-id:channel-id", gomock.Any()).Return(false, nil)
	test.storage.EXPECT().GetOne(ctx, "guild-id", "channel-id").Return(expect, nil)
	test.cache.EXPECT().SetWithJson(ctx, "autoreactions:guild-id:channel-id", expect, autoreactionsCacheTTL).Return(nil)

	actual, err := test.service.ObtainOne(ctx, "guild-id", "channel-id")
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestObtainOneNotFound() {
	ctx := context.Background()
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.cache.EXPECT().GetWithJson(ctx, "autoreactions:guild-id:channel-id", gomock.Any()).Return(false, nil)
	test.storage.EXPECT().GetOne(ctx, "guild-id", "channel-id").Return(nil, notFound)

	actual, err := test.service.ObtainOne(ctx, "guild-id", "channel-id")
	test.Nil(actual)
	test.ErrorIs(err, ErrAutoReactionNotFound)
}

func (test *test) TestObtainAll() {
	ctx := context.Background()
	expect := []domain.AutoReaction{
		{
			ID:        1,
			GuildID:   "guild-id",
			ChannelID: "channel-id",
			Type:      domain.AutoReactionTypeMessage,
			Emoji:     "🤖",
		},
	}

	test.cache.EXPECT().GetWithJson(ctx, "autoreactions:guild:guild-id", gomock.Any()).Return(false, nil)
	test.storage.EXPECT().GetAll(ctx, "guild-id").Return(expect, nil)
	test.cache.EXPECT().SetWithJson(ctx, "autoreactions:guild:guild-id", expect, autoreactionsCacheTTL).Return(nil)

	actual, err := test.service.ObtainAll(ctx, "guild-id")
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestObtainAllNotFound() {
	ctx := context.Background()
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.cache.EXPECT().GetWithJson(ctx, "autoreactions:guild:guild-id", gomock.Any()).Return(false, nil)
	test.storage.EXPECT().GetAll(ctx, "guild-id").Return(nil, notFound)

	actual, err := test.service.ObtainAll(ctx, "guild-id")
	test.Nil(actual)
	test.ErrorIs(err, ErrAutoReactionNotFound)
}

func (test *test) TestUpdateOne() {
	ctx := context.Background()
	expect := &domain.AutoReaction{
		ID:        1,
		GuildID:   "guild-id",
		ChannelID: "channel-id",
		Type:      domain.AutoReactionTypeMessage,
		Emoji:     "🤖",
	}

	test.storage.EXPECT().UpdateOne(ctx, *expect).Return(expect, nil)
	test.cache.EXPECT().Del(ctx, "autoreactions:guild-id:channel-id", "autoreactions:guild:guild-id").Return(nil)

	actual, err := test.service.UpdateOne(ctx, *expect)
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestUpdateOneNotFound() {
	ctx := context.Background()
	autoreaction := domain.AutoReaction{GuildID: "guild-id", ChannelID: "channel-id"}
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.storage.EXPECT().UpdateOne(ctx, autoreaction).Return(nil, notFound)

	actual, err := test.service.UpdateOne(ctx, autoreaction)
	test.Nil(actual)
	test.ErrorIs(err, ErrAutoReactionNotFound)
}

func (test *test) TestDeleteOne() {
	ctx := context.Background()
	expect := &domain.AutoReaction{
		ID:        1,
		GuildID:   "guild-id",
		ChannelID: "channel-id",
		Type:      domain.AutoReactionTypeMessage,
		Emoji:     "🤖",
	}

	test.storage.EXPECT().DeleteOne(ctx, "guild-id", "channel-id").Return(expect, nil)
	test.cache.EXPECT().Del(ctx, "autoreactions:guild-id:channel-id", "autoreactions:guild:guild-id").Return(nil)

	actual, err := test.service.DeleteOne(ctx, "guild-id", "channel-id")
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestDeleteOneNotFound() {
	ctx := context.Background()
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.storage.EXPECT().DeleteOne(ctx, "guild-id", "channel-id").Return(nil, notFound)

	actual, err := test.service.DeleteOne(ctx, "guild-id", "channel-id")
	test.Nil(actual)
	test.ErrorIs(err, ErrAutoReactionNotFound)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
