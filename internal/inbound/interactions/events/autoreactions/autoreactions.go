package autoreactions

import (
	"log/slog"

	"github.com/discord-go/discord.go/bot"
	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type Params struct {
	fx.In
	AutoReactionsService autoreactions.Interface
	DiscordClient        *discord.DiscordClient
}

type Event struct {
	logger               *slog.Logger
	autoReactionsService autoreactions.Interface
	discordClient        *discord.DiscordClient
}

func New(params Params) {
	event := Event{
		autoReactionsService: params.AutoReactionsService,
		discordClient:        params.DiscordClient,
	}

	params.DiscordClient.Bot.OnMessageCreate(func(ctx *bot.MessageContext) {
		event.AutoReactionsMessageCreate(ctx)
	})
}

func (command *Event) AutoReactionsMessageCreate(ctx *bot.MessageContext) {
	if ctx.Author.Bot {
		return
	}

	data, err := command.autoReactionsService.ObtainOne(ctx.Context(), ctx.GuildID.String(), ctx.ChannelID().String())
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
