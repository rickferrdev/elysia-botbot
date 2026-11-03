package templates

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/messages"
	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtemplates"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type MenuID struct {
	interaction *bot.InteractionContext
	id          int
}

type (
	Service struct {
		templates dbtemplates.Interface
	}

	FxArgs struct {
		fx.In

		TemplatesDB dbtemplates.Interface
	}
)

type Interface interface {
	ObtainOne(ctx context.Context, id int) (*domain.Template, error)
	DeleteOne(ctx context.Context, id int) (*domain.Template, error)
	UpdateOne(ctx context.Context, embed domain.Template) (*domain.Template, error)
	ObtainAll(ctx context.Context, guildID, authorID string) ([]domain.Template, error)
	HandlerMessageContent(ctx context.Context, id int, guildID, authorID, content string) (*domain.Template, error)
	HandlerAttachments(ctx context.Context, id int, guildID, authorID string, message messages.Attachment) (*domain.Template, error)
}

func New(FxArgs FxArgs) *Service {
	service := Service{
		templates: FxArgs.TemplatesDB,
	}

	return &service
}

func (service *Service) ObtainOne(ctx context.Context, id int) (*domain.Template, error) {
	result, err := service.templates.GetOne(ctx, id)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	return result, err
}

func (service *Service) DeleteOne(ctx context.Context, id int) (*domain.Template, error) {
	result, err := service.templates.DeleteOne(ctx, id)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	return result, err
}

func (service *Service) UpdateOne(ctx context.Context, embed domain.Template) (*domain.Template, error) {
	result, err := service.templates.UpdateOne(ctx, embed)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
		}
		if core.IsCode(err, repoerrors.CodeSqlConflict) {
			return nil, fmt.Errorf("%w: %w", ErrConflict, err)
		}
		return nil, err
	}

	return result, nil
}

func (service *Service) HandlerAttachments(ctx context.Context, id int, guildID, authorID string, message messages.Attachment) (*domain.Template, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, message.URL, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout: 15 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBadRequest, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %w", ErrBadRequest, err)
	}

	maxSize := 5 << 20
	data, err := io.ReadAll(
		io.LimitReader(response.Body, int64(maxSize)+1),
	)
	if err != nil {
		return nil, err
	}

	if len(data) > maxSize {
		return nil, ErrMaxSize
	}

	return service.HandlerMessageContent(ctx, id, guildID, authorID, strings.ToValidUTF8(string(data), "\uFFFD"))
}

func (service *Service) HandlerMessageContent(ctx context.Context, id int, guildID, authorID, content string) (*domain.Template, error) {
	formatted := strings.TrimSpace(content)
	if !json.Valid([]byte(formatted)) {
		return nil, ErrInvalidInput
	}

	var embedContent domain.EmbedContentDecode
	if err := json.Unmarshal([]byte(formatted), &embedContent); err != nil {
		return nil, ErrInvalidInput
	}

	embeds, err := json.Marshal(embedContent.Embeds)
	if err != nil {
		return nil, ErrInvalidInput
	}
	attachments, err := json.Marshal(embedContent.Attachments)
	if err != nil {
		return nil, ErrInvalidInput
	}

	now := time.Now()
	embed := domain.Template{
		ID:        id,
		GuildID:   guildID,
		AuthorID:  authorID,
		CreatedAt: now,
		UpdatedAt: now,
		TemplateContent: &domain.TemplateContent{
			Content:     embedContent.Content,
			Embeds:      string(embeds),
			Attachments: string(attachments),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	data, err := &domain.Template{}, nil
	if id > 0 {
		data, err = service.UpdateOne(ctx, embed)
	} else {
		data, err = service.templates.CreateOne(ctx, embed)
	}
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlConflict) {
			return nil, fmt.Errorf("%w: %w", ErrConflict, err)
		}
		return nil, err
	}

	return data, nil
}

func (service *Service) ObtainAll(ctx context.Context, guildID, authorID string) ([]domain.Template, error) {
	result, err := service.templates.GetAll(ctx, guildID, authorID)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return result, nil
}
