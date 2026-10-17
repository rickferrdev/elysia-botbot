package autoreactions

import (
	"log/slog"

	"github.com/discord-go/discord.go/bot"
	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"github.com/rickferrdev/elysia-botbot/internal/infra/redisclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type FxArgs struct {
	fx.In

	RedisClient   *redisclient.RedisClient
	DiscordClient *discordclient.DiscordClient
	AutoReactions autoreactions.Interface
}

type Event struct {
	redis         redisclient.RedisClient
	client        *discordclient.DiscordClient
	logger        *slog.Logger
	autoreactions autoreactions.Interface
}

func New(FxArgs FxArgs) {
	event := Event{
		autoreactions: FxArgs.AutoReactions,
		client:        FxArgs.DiscordClient,
	}

	FxArgs.DiscordClient.Bot.OnMessageCreate(func(ctx *bot.MessageContext) {
		event.AutoReactionsMessageCreate(ctx)
	})
}

func (command *Event) AutoReactionsMessageCreate(ctx *bot.MessageContext) {
	if ctx.Author.Bot {
		return
	}

	data, err := command.autoreactions.ObtainOne(ctx.Context(), ctx.GuildID.String(), ctx.ChannelID().String())
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			return
		}

		command.logger.Error("failed to get autoreaction", slog.String("error", err.Error()))
		return
	}

	switch data.Type {
	case domain.AutoReactionTypeAttachment:
		if len(ctx.Attachments) == 0 {
			return
		}

		if err := ctx.React(data.Emoji); err != nil {
			command.logger.Error(err.Error())
		}
		return

	case domain.AutoReactionTypeMessage:
		if len(ctx.Content) == 0 {
			return
		}

		if err := ctx.React(data.Emoji); err != nil {
			command.logger.Error(err.Error())
		}
		return

	default:
		return
	}
}
