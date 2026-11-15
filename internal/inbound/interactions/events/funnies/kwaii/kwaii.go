package kwaii

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/funnies/kwaii"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type FxArgs struct {
	fx.In

	KwaiiService  kwaii.Interface
	Env           *env.Env
	Logger        *slog.Logger
	DiscordClient *discordclient.DiscordClient
}

type webhook struct {
	id    snowflake.ID
	token string
}

type Event struct {
	service   kwaii.Interface
	client    *discordclient.DiscordClient
	logger    *slog.Logger
	webhookMu sync.Mutex
	webhooks  map[snowflake.ID]webhook
}

func New(args FxArgs) {
	event := &Event{
		service:  args.KwaiiService,
		client:   args.DiscordClient,
		logger:   args.Logger,
		webhooks: make(map[snowflake.ID]webhook),
	}

	args.DiscordClient.Bot.OnMessageCreate(event.MessageCreate)
}

func (event *Event) MessageCreate(ctx *bot.MessageContext) {
	if ctx.Author == nil || ctx.Author.Bot || strings.TrimSpace(ctx.Content) == "" || len(ctx.Attachments) > 0 {
		return
	}

	_, err := event.service.GetOne(ctx.Context(), ctx.Author.ID.String())
	if err != nil {
		if errors.Is(err, kwaii.ErrKwaiiEffectNotFound) {
			return
		}
		event.logger.Info("failed to retrieve from the database using user_id", slog.Any("error", err))
		return
	}

	requestCtx, cancel := context.WithTimeout(ctx.Context(), 30*time.Second)
	defer cancel()

	content, err := event.service.MakeKwaii(requestCtx, ctx.Content)
	if err != nil {
		event.logger.Error("failed to transform message with Gemini", slog.Any("error", err))
		return
	}
	if content == "" {
		event.logger.Error("gemini returned an empty transformed message")
		return
	}

	channelWebhook, err := event.webhookForChannel(requestCtx, ctx.ChannelID())
	if err != nil {
		event.logger.Error("failed to get Discord webhook", slog.Any("error", err))
		return
	}

	username := ctx.Author.Username
	if ctx.Author.GlobalName != nil && *ctx.Author.GlobalName != "" {
		username = *ctx.Author.GlobalName
	}

	_, err = event.client.Bot.Rest.ExecuteWebhook(requestCtx, channelWebhook.id, channelWebhook.token, rest.ExecuteWebhookParams{
		Content:         content,
		Username:        username,
		AvatarURL:       ctx.Author.DisplayAvatarURL(users.AvatarURLOptions{Size: 128}),
		AllowedMentions: &messages.AllowedMentions{},
	})
	if err != nil {
		event.logger.Error("failed to send transformed message via Discord webhook", slog.Any("error", err))
		return
	}

	if err := ctx.Delete(); err != nil {
		event.logger.Error("failed to delete original message", slog.Any("error", err))
	}
}

func (event *Event) webhookForChannel(ctx context.Context, channelID snowflake.ID) (webhook, error) {
	event.webhookMu.Lock()
	defer event.webhookMu.Unlock()

	if existing, ok := event.webhooks[channelID]; ok {
		return existing, nil
	}

	created, err := event.client.Bot.Rest.CreateWebhook(ctx, channelID, rest.CreateWebhookParams{
		Name: "Kwaii",
	})
	if err != nil {
		return webhook{}, fmt.Errorf("create webhook for channel %s: %w", channelID, err)
	}
	if created.Token == "" {
		return webhook{}, fmt.Errorf("Discord returned a webhook without a token for channel %s", channelID)
	}

	channelWebhook := webhook{id: created.ID, token: created.Token}
	event.webhooks[channelID] = channelWebhook
	return channelWebhook, nil
}
