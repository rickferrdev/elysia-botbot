package autoreactions

import (
	"log/slog"

	"github.com/discord-go/discord.go/bot"
	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type FxArgs struct {
	fx.In

	Logger        *slog.Logger
	DiscordClient *discordclient.DiscordClient
	AutoReactions autoreactions.Interface
}

type Event struct {
	client        *discordclient.DiscordClient
	logger        *slog.Logger
	autoreactions autoreactions.Interface
}

func New(args FxArgs) {
	event := Event{
		logger:        args.Logger,
		autoreactions: args.AutoReactions,
		client:        args.DiscordClient,
	}

	args.DiscordClient.Bot.OnMessageCreate(func(ctx *bot.MessageContext) {
		event.AutoReactionsMessageCreate(ctx)
	})
}

func (event *Event) AutoReactionsMessageCreate(ctx *bot.MessageContext) {
	if ctx.Author.Bot {
		return
	}

	data, err := event.autoreactions.ObtainOne(ctx.Context(), ctx.GuildID.String(), ctx.ChannelID().String())
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return
		}

		event.logger.Error(
			"failed to get auto-reaction",
			slog.String("guild_id", ctx.GuildID.String()),
			slog.String("channel_id", ctx.ChannelID().String()),
			slog.Any("error", err),
		)
		return
	}

	switch data.Type {
	case domain.AutoReactionTypeAttachment:
		if len(ctx.Attachments) == 0 {
			return
		}

		if err := ctx.React(data.Emoji); err != nil {
			event.logger.Error(
				"failed to add auto-reaction",
				slog.String("guild_id", ctx.GuildID.String()),
				slog.String("channel_id", ctx.ChannelID().String()),
				slog.Any("error", err),
			)
		}
		return

	case domain.AutoReactionTypeMessage:
		if len(ctx.Content) == 0 {
			return
		}

		if err := ctx.React(data.Emoji); err != nil {
			event.logger.Error(
				"failed to add auto-reaction",
				slog.String("guild_id", ctx.GuildID.String()),
				slog.String("channel_id", ctx.ChannelID().String()),
				slog.Any("error", err),
			)
		}
		return

	default:
		return
	}
}
