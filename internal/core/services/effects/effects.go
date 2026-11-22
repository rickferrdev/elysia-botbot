package effects

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbeffects"
	"github.com/rickferrdev/elysia-botbot/internal/platform/cache"
	"github.com/rickferrdev/elysia-botbot/internal/platform/genai"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

//go:embed prompts/rewrite.pt-BR.txt
var prompt string

const (
	Temperature     = 1.0
	MaxOutputTokens = 180
)

var (
	ErrEffectNotFound          = errors.New("effect not found")
	ErrEffectAlreadyExists     = errors.New("effect already exists")
	ErrStorageOperationFailed  = errors.New("storage operation failed")
	ErrGenerateOperationFailed = errors.New("generate operation failed")
)

type (
	Service struct {
		env     *env.Env
		http    *http.Client
		cache   cache.Interface
		genai   genai.Interface
		logger  *slog.Logger
		storage dbeffects.Interface
	}

	FxArgs struct {
		fx.In

		Env     *env.Env
		Genai   genai.Interface
		Cache   cache.Interface
		Logger  *slog.Logger
		DBKwaii dbeffects.Interface
	}
)

type Interface interface {
	Generate(ctx context.Context, content string) (string, error)

	ObtainAll(ctx context.Context, guildID string) (*[]domain.Effect, error)
	ObtainOne(ctx context.Context, guildID, userID string) (*domain.Effect, error)
	CreateOne(ctx context.Context, effect domain.Effect) (*domain.Effect, error)
	UpdateOne(ctx context.Context, effect domain.Effect) (*domain.Effect, error)
	DeleteOne(ctx context.Context, guildID, userID string) (*domain.Effect, error)
}

func New(args FxArgs) *Service {
	service := Service{
		env:     args.Env,
		genai:   args.Genai,
		cache:   args.Cache,
		storage: args.DBKwaii,
		logger:  args.Logger,
		http:    &http.Client{Timeout: 30 * time.Second},
	}

	return &service
}

const effectCacheTTL = time.Minute * 10

func (service *Service) getCacheKey(guildID, userID string) string {
	return fmt.Sprintf("effect:%v:%v", guildID, userID)
}

func (service *Service) getAllCacheKey(guildID string) string {
	return fmt.Sprintf("effect:%v", guildID)
}

func (service *Service) ObtainOne(ctx context.Context, guildID, userID string) (*domain.Effect, error) {
	var cached domain.Effect
	if found, err := service.cache.GetWithJson(
		ctx,
		service.getCacheKey(guildID, userID),
		&cached,
	); err == nil && found {
		if cached.GuildID == guildID && cached.UserID == userID {
			return &cached, nil
		}
	}
	data, err := service.storage.GetOne(ctx, guildID, userID)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, ErrEffectNotFound
		}
		return nil, ErrStorageOperationFailed
	}

	err = service.cache.SetWithJson(
		ctx,
		service.getCacheKey(guildID, userID),
		data,
		effectCacheTTL,
	)
	if err != nil {
		service.logger.Error(
			"failed to cache user_id",
			slog.String("guild_id", guildID),
			slog.String("user_id", userID),
			slog.Any("error", err),
		)
	}

	return data, nil
}

func (service *Service) ObtainAll(ctx context.Context, guildID string) (*[]domain.Effect, error) {
	var cached []domain.Effect
	if found, err := service.cache.GetWithJson(
		ctx,
		service.getAllCacheKey(guildID),
		&cached,
	); err == nil && found {
		return &cached, nil
	}

	data, err := service.storage.GetAll(ctx, guildID, 100)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, ErrEffectNotFound
		}
		return nil, ErrStorageOperationFailed
	}

	err = service.cache.SetWithJson(
		ctx,
		service.getAllCacheKey(guildID),
		data,
		effectCacheTTL,
	)
	if err != nil {
		service.logger.Error(
			"failed to cache auto-reactions",
			slog.String("guild_id", guildID),
			slog.Any("error", err),
		)
	}
	return data, nil
}

func (service *Service) CreateOne(ctx context.Context, effect domain.Effect) (*domain.Effect, error) {
	data, err := service.storage.CreateOne(ctx, effect)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlConflict) {
			return nil, ErrEffectAlreadyExists
		}
		return nil, ErrStorageOperationFailed
	}

	err = service.cache.Del(
		ctx,
		service.getCacheKey(data.GuildID, data.UserID),
		service.getAllCacheKey(data.GuildID),
	)
	if err != nil {
		service.logger.Error(
			"failed to invalidate effect cache",
			slog.String("operation", "create"),
			slog.String("guild_id", data.GuildID),
			slog.String("user_id", data.UserID),
			slog.Any("error", err),
		)
	}

	return data, nil
}

func (service *Service) UpdateOne(ctx context.Context, effect domain.Effect) (*domain.Effect, error) {
	data, err := service.storage.UpdateOne(ctx, effect)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, ErrEffectNotFound
		}
		return nil, ErrStorageOperationFailed
	}

	err = service.cache.Del(
		ctx,
		service.getCacheKey(effect.GuildID, effect.UserID),
		service.getAllCacheKey(effect.GuildID),
	)
	if err != nil {
		service.logger.Error(
			"failed to invalidate effect cache",
			slog.String("operation", "update"),
			slog.String("guild_id", effect.GuildID),
			slog.String("user_id", effect.UserID),
			slog.Any("error", err),
		)
	}
	return data, nil
}

func (service *Service) DeleteOne(ctx context.Context, guildID, userID string) (*domain.Effect, error) {
	data, err := service.storage.DeleteOne(ctx, userID)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return nil, ErrEffectNotFound
		}
		return nil, ErrStorageOperationFailed
	}

	err = service.cache.Del(
		ctx,
		service.getCacheKey(guildID, userID),
		service.getAllCacheKey(guildID),
	)
	if err != nil {
		service.logger.Error(
			"failed to invalidate effect cache",
			slog.String("operation", "update"),
			slog.String("guild_id", guildID),
			slog.String("user_id", userID),
			slog.Any("error", err),
		)
	}
	return data, nil
}

func (service *Service) Generate(ctx context.Context, content string) (string, error) {
	data, err := service.genai.GenerateContent(ctx, content, genai.GenerateContentConfig{
		SystemInstruction: prompt,
		Temperature:       1.0,
		MaxOutputTokens:   MaxOutputTokens,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGenerateOperationFailed, err)
	}
	if data == nil {
		return "", ErrGenerateOperationFailed
	}

	if strings.TrimSpace(data.Text) == "" {
		return "", ErrGenerateOperationFailed
	}

	return data.Text, nil
}
