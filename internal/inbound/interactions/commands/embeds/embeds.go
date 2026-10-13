package embeds

import (
	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/permissions"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/embeds"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type Params struct {
	fx.In
	EmbedsService embeds.Interface
	Client        *discord.DiscordClient
}

type Embeds struct {
	service embeds.Interface
	client  *discord.DiscordClient
}

func New(params Params) *Embeds {
	embeds := Embeds{
		client:  params.Client,
		service: params.EmbedsService,
	}

	params.Client.Router.Command(
		"embeds",
		"Configurações úteis relacionadas a embeds.",
		embeds.execute,
	).
		Use(
			bot.RequireBotPermissions(
				permissions.Build(
					permissions.ManageMessages,
					permissions.AddReactions,
					permissions.ReadMessageHistory,
				),
			),
		).
		Use(bot.RequirePermissions(
			permissions.Build(
				permissions.Administrator,
			),
		))
	return &embeds
}

func (embeds *Embeds) execute(ctx *bot.InteractionContext) {
	embeds.service.EmbedsSetup(ctx)
}
