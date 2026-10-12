package autoreactions

import (
	"github.com/discord-go/discord.go/bot"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type Params struct {
	fx.In
	AutoReactionsService autoreactions.Interface
	DiscordClient        *discord.DiscordClient
}

func New(params Params) {
	params.DiscordClient.Bot.OnMessageCreate(func(ctx *bot.MessageContext) {
		params.AutoReactionsService.AutoReactionsMessageCreate(ctx)
	})
}
