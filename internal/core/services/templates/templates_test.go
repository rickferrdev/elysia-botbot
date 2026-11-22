package templates

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/messages"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	mock_templates "github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtemplates/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type test struct {
	suite.Suite
	storage *mock_templates.MockInterface
	service *Service
}

func (test *test) SetupTest() {
	test.storage = mock_templates.NewMockInterface(gomock.NewController(test.T()))
	test.service = New(FxArgs{TemplatesDB: test.storage})
}

func (test *test) TestObtainOne() {
	ctx := context.Background()
	expect := &domain.Template{ID: 1, GuildID: "guild-id", AuthorID: "author-id"}

	test.storage.EXPECT().GetOne(ctx, 1).Return(expect, nil)

	actual, err := test.service.ObtainOne(ctx, 1)
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestObtainOneNotFound() {
	ctx := context.Background()
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.storage.EXPECT().GetOne(ctx, 1).Return(nil, notFound)

	actual, err := test.service.ObtainOne(ctx, 1)
	test.Nil(actual)
	test.ErrorIs(err, ErrTemplateNotFound)
}

func (test *test) TestDeleteOne() {
	ctx := context.Background()
	expect := &domain.Template{ID: 1, GuildID: "guild-id", AuthorID: "author-id"}

	test.storage.EXPECT().DeleteOne(ctx, 1).Return(expect, nil)

	actual, err := test.service.DeleteOne(ctx, 1)
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestDeleteOneNotFound() {
	ctx := context.Background()
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.storage.EXPECT().DeleteOne(ctx, 1).Return(nil, notFound)

	actual, err := test.service.DeleteOne(ctx, 1)
	test.Nil(actual)
	test.ErrorIs(err, ErrTemplateNotFound)
}

func (test *test) TestUpdateOne() {
	ctx := context.Background()
	template := domain.Template{ID: 1, GuildID: "guild-id", AuthorID: "author-id"}
	expect := &template

	test.storage.EXPECT().UpdateOne(ctx, template).Return(expect, nil)

	actual, err := test.service.UpdateOne(ctx, template)
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestUpdateOneConflict() {
	ctx := context.Background()
	template := domain.Template{ID: 1, GuildID: "guild-id", AuthorID: "author-id"}
	conflict := repoerrors.NewErrorSqlConflict(errors.New("conflict"))

	test.storage.EXPECT().UpdateOne(ctx, template).Return(nil, conflict)

	actual, err := test.service.UpdateOne(ctx, template)
	test.Nil(actual)
	test.ErrorIs(err, ErrTemplateAlreadyExists)
}

func (test *test) TestObtainAll() {
	ctx := context.Background()
	expect := []domain.Template{{ID: 1, GuildID: "guild-id", AuthorID: "author-id"}}

	test.storage.EXPECT().GetAll(ctx, "guild-id", "author-id").Return(expect, nil)

	actual, err := test.service.ObtainAll(ctx, "guild-id", "author-id")
	test.NoError(err)
	test.Equal(expect, actual)
}

func (test *test) TestObtainAllNotFound() {
	ctx := context.Background()
	notFound := repoerrors.NewErrorSqlNoRows(sql.ErrNoRows)

	test.storage.EXPECT().GetAll(ctx, "guild-id", "author-id").Return(nil, notFound)

	actual, err := test.service.ObtainAll(ctx, "guild-id", "author-id")
	test.Nil(actual)
	test.ErrorIs(err, ErrTemplateNotFound)
}

func (test *test) TestHandlerMessageContent() {
	ctx := context.Background()
	var persisted domain.Template

	test.storage.EXPECT().CreateOne(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, template domain.Template) (*domain.Template, error) {
			persisted = template
			return &template, nil
		},
	)

	actual, err := test.service.HandlerMessageContent(
		ctx,
		0,
		"guild-id",
		"author-id",
		`{"content":"hello","embeds":[],"attachments":[]}`,
	)
	test.NoError(err)
	test.Equal("hello", persisted.TemplateContent.Content)
	test.JSONEq("[]", persisted.TemplateContent.Embeds)
	test.JSONEq("[]", persisted.TemplateContent.Attachments)
	test.Equal(persisted.CreatedAt, persisted.TemplateContent.CreatedAt)
	test.Equal(persisted.UpdatedAt, persisted.TemplateContent.UpdatedAt)
	test.NotZero(persisted.CreatedAt)
	test.Equal(&persisted, actual)
}

func (test *test) TestHandlerMessageContentInvalidInput() {
	actual, err := test.service.HandlerMessageContent(
		context.Background(),
		0,
		"guild-id",
		"author-id",
		"not-json",
	)
	test.Nil(actual)
	test.ErrorIs(err, ErrInvalidTemplateInput)
}

func (test *test) TestHandlerAttachments() {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"content":"from attachment","embeds":[],"attachments":[]}`))
	}))
	defer server.Close()

	var persisted domain.Template
	test.storage.EXPECT().CreateOne(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, template domain.Template) (*domain.Template, error) {
			persisted = template
			return &template, nil
		},
	)

	actual, err := test.service.HandlerAttachments(
		ctx,
		0,
		"guild-id",
		"author-id",
		messages.Attachment{URL: server.URL},
	)
	test.NoError(err)
	test.Equal("from attachment", persisted.TemplateContent.Content)
	test.Equal(&persisted, actual)
}

func (test *test) TestHandlerAttachmentsDownloadFailure() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	actual, err := test.service.HandlerAttachments(
		context.Background(),
		0,
		"guild-id",
		"author-id",
		messages.Attachment{URL: server.URL},
	)
	test.Nil(actual)
	test.ErrorIs(err, ErrTemplateAttachmentDownloadFailed)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
