package effects

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	mock_effects "github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbeffects/mocks"
	mock_cache "github.com/rickferrdev/elysia-botbot/internal/platform/cache/mocks"
	"github.com/rickferrdev/elysia-botbot/internal/platform/genai"
	mock_genai "github.com/rickferrdev/elysia-botbot/internal/platform/genai/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type test struct {
	suite.Suite
	storage *mock_effects.MockInterface
	cache   *mock_cache.MockInterface
	genai   *mock_genai.MockInterface
	service *Service
}

func (test *test) SetupTest() {
	controller := gomock.NewController(test.T())
	test.storage = mock_effects.NewMockInterface(controller)
	test.cache = mock_cache.NewMockInterface(controller)
	test.genai = mock_genai.NewMockInterface(controller)
	test.service = New(FxArgs{
		Env:     &env.Env{},
		Genai:   test.genai,
		Cache:   test.cache,
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		DBKwaii: test.storage,
	})
}

func (test *test) TestObtainOne() {
	ctx := context.Background()
	expect := &domain.Effect{ID: 1, GuildID: "guild-id", UserID: "user-id", Type: domain.EffectTypeKwaii}

	test.cache.EXPECT().GetWithJson(ctx, "effect:guild-id:user-id", gomock.Any()).Return(false, nil)
	test.storage.EXPECT().GetOne(ctx, "guild-id", "user-id").Return(expect, nil)
	test.cache.EXPECT().SetWithJson(ctx, "effect:guild-id:user-id", expect, effectCacheTTL).Return(nil)

	actual, err := test.service.ObtainOne(ctx, "guild-id", "user-id")
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestObtainOneNotFound() {
	ctx := context.Background()
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.cache.EXPECT().GetWithJson(ctx, "effect:guild-id:user-id", gomock.Any()).Return(false, nil)
	test.storage.EXPECT().GetOne(ctx, "guild-id", "user-id").Return(nil, notFound)

	actual, err := test.service.ObtainOne(ctx, "guild-id", "user-id")
	test.Nil(actual)
	test.ErrorIs(err, ErrEffectNotFound)
}

func (test *test) TestObtainAll() {
	ctx := context.Background()
	expect := []domain.Effect{{ID: 1, GuildID: "guild-id", UserID: "user-id", Type: domain.EffectTypeKwaii}}

	test.cache.EXPECT().GetWithJson(ctx, "effect:guild-id", gomock.Any()).Return(false, nil)
	test.storage.EXPECT().GetAll(ctx, "guild-id", 100).Return(&expect, nil)
	test.cache.EXPECT().SetWithJson(ctx, "effect:guild-id", &expect, effectCacheTTL).Return(nil)

	actual, err := test.service.ObtainAll(ctx, "guild-id")
	test.NoError(err)
	test.Equal(&expect, actual)
}

func (test *test) TestObtainAllNotFound() {
	ctx := context.Background()
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.cache.EXPECT().GetWithJson(ctx, "effect:guild-id", gomock.Any()).Return(false, nil)
	test.storage.EXPECT().GetAll(ctx, "guild-id", 100).Return(nil, notFound)

	actual, err := test.service.ObtainAll(ctx, "guild-id")
	test.Nil(actual)
	test.ErrorIs(err, ErrEffectNotFound)
}

func (test *test) TestCreateOne() {
	ctx := context.Background()
	expect := &domain.Effect{ID: 1, GuildID: "guild-id", UserID: "user-id", Type: domain.EffectTypeKwaii}

	test.storage.EXPECT().CreateOne(ctx, *expect).Return(expect, nil)
	test.cache.EXPECT().Del(ctx, "effect:guild-id:user-id", "effect:guild-id").Return(nil)

	actual, err := test.service.CreateOne(ctx, *expect)
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestCreateOneConflict() {
	ctx := context.Background()
	effect := domain.Effect{GuildID: "guild-id", UserID: "user-id"}
	conflict := repoerrors.NewErrorSqlConflict(errors.New("conflict"))

	test.storage.EXPECT().CreateOne(ctx, effect).Return(nil, conflict)

	actual, err := test.service.CreateOne(ctx, effect)
	test.Nil(actual)
	test.ErrorIs(err, ErrEffectAlreadyExists)
}

func (test *test) TestUpdateOne() {
	ctx := context.Background()
	effect := domain.Effect{ID: 1, GuildID: "guild-id", UserID: "user-id", Type: domain.EffectTypeKwaii}
	expect := &effect

	test.storage.EXPECT().UpdateOne(ctx, effect).Return(expect, nil)
	test.cache.EXPECT().Del(ctx, "effect:guild-id:user-id", "effect:guild-id").Return(nil)

	actual, err := test.service.UpdateOne(ctx, effect)
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestUpdateOneNotFound() {
	ctx := context.Background()
	effect := domain.Effect{GuildID: "guild-id", UserID: "user-id"}
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.storage.EXPECT().UpdateOne(ctx, effect).Return(nil, notFound)

	actual, err := test.service.UpdateOne(ctx, effect)
	test.Nil(actual)
	test.ErrorIs(err, ErrEffectNotFound)
}

func (test *test) TestDeleteOne() {
	ctx := context.Background()
	expect := &domain.Effect{ID: 1, GuildID: "guild-id", UserID: "user-id", Type: domain.EffectTypeKwaii}

	test.storage.EXPECT().DeleteOne(ctx, "user-id").Return(expect, nil)
	test.cache.EXPECT().Del(ctx, "effect:guild-id:user-id", "effect:guild-id").Return(nil)

	actual, err := test.service.DeleteOne(ctx, "guild-id", "user-id")
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestDeleteOneNotFound() {
	ctx := context.Background()
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.storage.EXPECT().DeleteOne(ctx, "user-id").Return(nil, notFound)

	actual, err := test.service.DeleteOne(ctx, "guild-id", "user-id")
	test.Nil(actual)
	test.ErrorIs(err, ErrEffectNotFound)
}

func (test *test) TestGenerate() {
	ctx := context.Background()

	prompt, err := test.service.CollectPrompt(domain.EffectTypeKwaii)
	test.NoError(err)
	test.NotEmpty(prompt)

	expect := &genai.GenerateContentResponse{Text: "rewritten"}
	config := genai.GenerateContentConfig{
		SystemInstruction: prompt,
		Temperature:       Temperature,
		MaxOutputTokens:   MaxOutputTokens,
	}

	test.genai.EXPECT().GenerateContent(ctx, "input", config).Return(expect, nil)

	actual, err := test.service.Generate(ctx, "input", domain.EffectTypeKwaii)
	test.NoError(err)
	test.Equal(expect.Text, actual)
}

func (test *test) TestGenerateFailure() {
	ctx := context.Background()

	prompt, err := test.service.CollectPrompt(domain.EffectTypeKwaii)
	test.NoError(err)
	test.NotEmpty(prompt)

	generationErr := errors.New("generation failed")
	config := genai.GenerateContentConfig{
		SystemInstruction: prompt,
		Temperature:       Temperature,
		MaxOutputTokens:   MaxOutputTokens,
	}

	test.genai.EXPECT().GenerateContent(ctx, "input", config).Return(nil, generationErr)

	actual, err := test.service.Generate(ctx, "input", domain.EffectTypeKwaii)
	test.Empty(actual)
	test.ErrorIs(err, ErrGenerateOperationFailed)
	test.ErrorIs(err, generationErr)
}

func (test *test) TestGenerateEmptyResponse() {
	ctx := context.Background()

	prompt, err := test.service.CollectPrompt(domain.EffectTypeKwaii)
	test.NoError(err)
	test.NotEmpty(prompt)

	config := genai.GenerateContentConfig{
		SystemInstruction: prompt,
		Temperature:       Temperature,
		MaxOutputTokens:   MaxOutputTokens,
	}

	test.genai.EXPECT().GenerateContent(ctx, "input", config).Return(&genai.GenerateContentResponse{Text: "  "}, nil)

	actual, err := test.service.Generate(ctx, "input", domain.EffectTypeKwaii)
	test.Empty(actual)
	test.ErrorIs(err, ErrGenerateOperationFailed)
}

func (test *test) TestCollectPrompt() {
	expect, err := os.ReadFile("./prompts/kwaii.pt-BR.txt")
	test.NoError(err)
	workingDir, err := os.Getwd()
	test.NoError(err)
	test.NoError(os.Chdir(test.T().TempDir()))
	defer func() {
		if err := os.Chdir(workingDir); err != nil {
			test.T().Errorf("failed to restore working directory: %v", err)
		}
	}()

	actual, err := test.service.CollectPrompt(domain.EffectTypeKwaii)
	test.NoError(err)
	test.Equal(string(expect), actual)

	christmasPrompt, err := test.service.CollectPrompt(domain.EffectTypeChristmas)
	test.NoError(err)
	test.NotEmpty(christmasPrompt)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
