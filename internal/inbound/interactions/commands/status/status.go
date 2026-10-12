package status

import (
	"log/slog"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/users"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type Params struct {
	fx.In
	Client *discord.DiscordClient
	Logger *slog.Logger
}

type StatusCommand struct {
	client *discord.DiscordClient
	logger *slog.Logger
}

func New(params Params) *StatusCommand {
	status := StatusCommand{
		client: params.Client,
		logger: params.Logger,
	}

	params.Client.Router.Command(
		"status",
		"Obter a estado atual do bot.",
		status.execute,
	)
	return &status
}

func (status *StatusCommand) execute(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		status.logger.Error(err.Error())
		return
	}

	embed := messages.
		NewEmbedBuilder().
		SetTitle("୨୧ estado atual do bot! ♡").
		SetAuthor(
			"Elysia-BotBot",
			"https://github.com/rickferrdev/elysia-botbot",
			ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{}),
		).
		SetDescription("Se você recebeu esta mensagem, isso pode significar que o bot está totalmente operacional\n").
		SetThumbnail(ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xFFF1C9).
		SetTimestampNow()

	embedBuilded := embed.Build()
	if err := embedBuilded.Validate(); err != nil {
		status.logger.Error(err.Error())
		return
	}

	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Embeds: &[]messages.Embed{embedBuilded},
	}); err != nil {
		status.logger.Error(err.Error())
		return
	}
}
