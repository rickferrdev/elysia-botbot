package ping

import (
	"fmt"
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

type Ping struct {
	client *discord.DiscordClient
	logger *slog.Logger
}

func New(params Params) *Ping {
	ping := Ping{
		client: params.Client,
		logger: params.Logger,
	}

	params.Client.Router.Command(
		"ping",
		"Obter a latência atual do bot.",
		ping.execute,
	)
	return &ping
}

func (ping *Ping) execute(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		ping.logger.Error(err.Error())
		return
	}

	latency := ping.client.Bot.GatewayLatency()

	embed := messages.NewEmbedBuilder()
	build := embed.
		SetTitle("୨୧ pong! ♡").
		SetThumbnail(ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})).
		SetDescription(
			fmt.Sprintf("Tudo certinho por aqui, estrelinha!\n✦ latência: %dms", latency.Milliseconds()),
		).
		SetColor(0xFFF1C9).
		Build()
	if err := build.Validate(); err != nil {
		ping.logger.Error(err.Error())
		return
	}

	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Embeds: &[]messages.Embed{build},
	}); err != nil {
		ping.logger.Error(err.Error())
		return
	}
}
