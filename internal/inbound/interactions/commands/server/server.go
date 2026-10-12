package server

import (
	"log/slog"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/permissions"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/server"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type Params struct {
	fx.In
	ServerService server.Interface
	Client        *discord.DiscordClient
	Logger        *slog.Logger
}

type Server struct {
	service server.Interface
	client  *discord.DiscordClient
	logger  *slog.Logger
}

func New(args Params) *Server {
	server := Server{
		client:  args.Client,
		logger:  args.Logger,
		service: args.ServerService,
	}

	args.Client.Router.Command(
		"server",
		"Configurações úteis para o servidor.",
		server.execute,
	).
		Use(bot.RequirePermissions(
			permissions.Build(
				permissions.Administrator,
			),
		))
	return &server
}

func (server *Server) execute(ctx *bot.InteractionContext) {
	server.service.ServerSetup(ctx, server.client.Router)
}
