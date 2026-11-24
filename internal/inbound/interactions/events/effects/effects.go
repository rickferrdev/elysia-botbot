package effects

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/users"
	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/effects"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"github.com/rickferrdev/elysia-botbot/internal/infra/logger"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type FxArgs struct {
	fx.In

	KwaiiService  effects.Interface
	Env           *env.Env
	Logger        *slog.Logger
	DiscordClient *discordclient.DiscordClient
}

type Event struct {
	service effects.Interface
	client  *discordclient.DiscordClient
	logger  *slog.Logger
}

func New(args FxArgs) {
	event := &Event{
		service: args.KwaiiService,
		client:  args.DiscordClient,
		logger:  args.Logger,
	}

	args.DiscordClient.Bot.OnMessageCreate(event.MessageCreate)
}

func (event *Event) MessageCreate(ctx *bot.MessageContext) {
	if ctx.Author == nil || ctx.Author.Bot || strings.TrimSpace(ctx.Content) == "" || len(ctx.Attachments) > 0 {
		return
	}

	effect, err := event.service.ObtainOne(ctx.Context(), ctx.GuildID.String(), ctx.Author.ID.String())
	if err != nil {
		if errors.Is(err, effects.ErrEffectNotFound) {
			return
		}
		event.logger.Info(
			logger.EffectEventFailedRetrieve,
			slog.String("guild_id", ctx.GuildID.String()),
			slog.String("user_id", ctx.Author.ID.String()),
			slog.Any("error", err),
		)
		return
	}
	if effect.UserID != ctx.Author.ID.String() {
		return
	}
	if effect.Expiration.Before(time.Now()) {
		_, err := event.service.DeleteOne(ctx.Context(), ctx.GuildID.String(), effect.UserID)
		if err != nil {
			event.logger.Info(
				logger.EffectEventFailedDeleteExpired,
				slog.String("guild_id", ctx.GuildID.String()),
				slog.String("user_id", effect.UserID),
				slog.Any("error", err),
			)
			return
		}
	}

	timeout, cancel := context.WithTimeout(ctx.Context(), 30*time.Second)
	defer cancel()

	content, err := event.service.Generate(timeout, ctx.Content, effect.Type)
	if err != nil {
		event.logger.Error(logger.EffectEventFailedGenerateMessage, slog.Any("error", err))
		return
	}

	username := ctx.Author.Username
	if ctx.Author.GlobalName != nil && *ctx.Author.GlobalName != "" {
		username = *ctx.Author.GlobalName
	}

	avatar := ctx.Author.DisplayAvatarURL(users.AvatarURLOptions{})
	hook, err := event.client.Bot.Rest.CreateWebhook(ctx.Context(), ctx.ChannelID(), rest.CreateWebhookParams{
		Name:   "Effect",
		Avatar: &avatar,
	})
	if err != nil {
		event.logger.Error(
			logger.EffectEventFailedCreateWebhook,
			slog.String("channel_id", ctx.ChannelID().String()),
			slog.Any("error", err),
		)
		return
	}
	defer func() {
		if err := event.client.Bot.Rest.DeleteWebhook(ctx.Context(), hook.ID); err != nil {
			event.logger.Error(
				logger.EffectEventFailedDeleteWebhook,
				slog.String("channel_id", ctx.ChannelID().String()),
				slog.Any("error", err),
			)
		}
	}()

	_, err = event.client.Bot.Rest.ExecuteWebhook(timeout, hook.ID, hook.Token, rest.ExecuteWebhookParams{
		Content:   content,
		Username:  username,
		AvatarURL: avatar,
	})
	if err != nil {
		event.logger.Error(
			logger.EffectEventFailedSendWebhook,
			slog.String("channel_id", ctx.ChannelID().String()),
			slog.Any("error", err),
		)
		return
	}

	if err := ctx.Delete(); err != nil {
		event.logger.Error(
			logger.EffectEventFailedDeleteMessage,
			slog.String("channel_id", ctx.ChannelID().String()),
			slog.Any("error", err),
		)
	}
}
