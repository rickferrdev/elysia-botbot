package genai

import (
	"context"
	"errors"
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"github.com/rickferrdev/elysia-botbot/internal/core"
	"go.uber.org/fx"
	"google.golang.org/genai"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

const (
	CodeBadRequest     core.ErrorCode = "BAD_REQUEST"
	CodeUnauthorized   core.ErrorCode = "UNAUTHORIZED"
	CodeInternalServer core.ErrorCode = "INTERNAL_SERVER"
)

//go:generate go run go.uber.org/mock/mockgen -source=genai.go -destination=mocks/mock_genai.go -package=mock_genai
type (
	Platform struct {
		Client *genai.Client
	}

	FxArgs struct {
		fx.In
		Env *env.Env
	}

	Interface interface {
		GenerateContent(ctx context.Context, text string, config GenerateContentConfig) (*GenerateContentResponse, error)
	}
)

func New(args FxArgs) (*Platform, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: args.Env.GeminiAPIKey,
	})
	if err != nil {
		return nil, err
	}

	platform := Platform{
		Client: client,
	}
	return &platform, nil
}

type GenerateContentConfig struct {
	SystemInstruction string
	Temperature       float32
	MaxOutputTokens   int32
}

type GenerateContentResponse struct {
	Text       string
	CreateTime time.Time
}

const maxAttempts = 3

func (platform *Platform) GenerateContent(ctx context.Context, text string, config GenerateContentConfig) (*GenerateContentResponse, error) {
	if platform == nil || platform.Client == nil {
		return nil, core.NewAppError("genai client unavailable", CodeInternalServer, errors.New("genai client is not configured"))
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		response, err := platform.Client.Models.GenerateContent(ctx,
			"gemini-3.5-flash-lite",
			genai.Text(text),
			&genai.GenerateContentConfig{
				SystemInstruction: genai.Text(config.SystemInstruction)[0],
				Temperature:       &config.Temperature,
				MaxOutputTokens:   config.MaxOutputTokens,
			})
		if err == nil {
			if response == nil {
				return nil, core.NewAppError("genai returned an empty response", CodeInternalServer, errors.New("empty response"))
			}

			return &GenerateContentResponse{
				Text:       response.Text(),
				CreateTime: response.CreateTime,
			}, nil
		}

		if e, ok := errors.AsType[genai.APIError](err); ok {
			switch e.Code {
			case 400:
				return nil, core.NewAppError("invalid request", CodeBadRequest, err)
			case 401, 403:
				return nil, core.NewAppError("client unauthorized", CodeUnauthorized, err)
			}
		}

		if !isRetryableGenAIError(err) || attempt == maxAttempts {
			return nil, core.NewAppError("genai request failed", CodeInternalServer, err)
		}

		delay := time.Duration(attempt) * time.Second
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, core.NewAppError("genai request canceled", CodeInternalServer, ctx.Err())
		case <-timer.C:
		}
	}

	return nil, core.NewAppError("genai request failed", CodeInternalServer, errors.New("retry attempts exhausted"))
}

func isRetryableGenAIError(err error) bool {
	apiError, ok := errors.AsType[genai.APIError](err)
	if !ok {
		return false
	}

	switch apiError.Code {
	case 429, 503, 504:
		return true
	default:
		return false
	}
}
