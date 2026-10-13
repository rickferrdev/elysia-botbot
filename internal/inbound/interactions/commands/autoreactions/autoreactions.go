package autoreactions

import (
	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/permissions"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type Params struct {
	fx.In
	AutoReactionsService autoreactions.Interface
	Client               *discord.DiscordClient
}

type Server struct {
	service autoreactions.Interface
	client  *discord.DiscordClient
}

func New(params Params) *Server {
	server := Server{
		client:  params.Client,
		service: params.AutoReactionsService,
	}

	params.Client.Router.Command(
		"autoreactions",
		"Configurações úteis relacionadas ao autoreactions.",
		server.execute,
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
	return &server
}

func (server *Server) execute(ctx *bot.InteractionContext) {
	server.service.AutoReactionsSetup(ctx)
}
