package embeds

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
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/embeds"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type MenuID struct {
	interaction *bot.InteractionContext
	id          int
}

type (
	Service struct {
		database embeds.Interface
	}

	Params struct {
		fx.In
		EmbedsDatabase embeds.Interface
	}
)

type Interface interface {
	ObtainOne(ctx context.Context, id int) (*domain.Embed, error)
	DeleteOne(ctx context.Context, id int) (*domain.Embed, error)
	UpsertOne(ctx context.Context, embed domain.Embed) (*domain.Embed, error)
	ObtainAll(ctx context.Context, guildID, authorID string) ([]domain.Embed, error)
	HandlerMessageContent(ctx context.Context, id int, guildID, authorID, content string) (*domain.Embed, error)
	HandlerAttachments(ctx context.Context, id int, guildID, authorID string, message messages.Attachment) (*domain.Embed, error)
}

func New(params Params) *Service {
	service := Service{
		database: params.EmbedsDatabase,
	}

	return &service
}

func (service *Service) ObtainOne(ctx context.Context, id int) (*domain.Embed, error) {
	result, err := service.database.GetOne(ctx, id)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return result, err
}
func (service *Service) DeleteOne(ctx context.Context, id int) (*domain.Embed, error) {
	result, err := service.database.DeleteOne(ctx, id)
	if core.IsCode(err, repoerrors.CodeSqlNoRows) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return result, err
}
func (service *Service) UpsertOne(ctx context.Context, embed domain.Embed) (*domain.Embed, error) {
	result, err := service.database.UpsertOne(ctx, embed)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConflict, err)
	}
	return result, nil
}
func (service *Service) HandlerAttachments(ctx context.Context, id int, guildID, authorID string, message messages.Attachment) (*domain.Embed, error) {
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
func (service *Service) HandlerMessageContent(ctx context.Context, id int, guildID, authorID, content string) (*domain.Embed, error) {
	formatted := strings.TrimSpace(content)
	if !json.Valid([]byte(formatted)) {
		return nil, ErrInvalidInput
	}
	var embedContent domain.EmbedContentDecode
	if err := json.Unmarshal([]byte(formatted), &embedContent); err != nil {
		return nil, ErrInvalidInput
	}
	bytes, err := json.Marshal(embedContent)
	if err != nil {
		return nil, ErrInvalidInput
	}
	embed := domain.Embed{
		ID:        id,
		GuildID:   guildID,
		UserID:    authorID,
		Content:   string(bytes),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	data, err := &domain.Embed{}, nil
	if id > 0 {
		data, err = service.database.UpsertOne(ctx, embed)
	} else {
		data, err = service.database.CreateOne(ctx, embed)
	}
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlConflict) {
			return nil, ErrConflict
		}
		return nil, err
	}
	return data, nil
}
func (service *Service) ObtainAll(ctx context.Context, guildID, authorID string) ([]domain.Embed, error) {
	result, err := service.database.GetAll(ctx, guildID, authorID)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return result, nil
}
