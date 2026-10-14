package server

import (
	"errors"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/permissions"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/server"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type Args struct {
	fx.In
	DiscordClient    *discord.DiscordClient
	ResponderService responder.Interface
	ServerService    server.Interface
	InteractionFlow  interactionflow.Interface
}

type Command struct {
	interactionFlow  interactionflow.Interface
	discordClient    *discord.DiscordClient
	serverService    server.Interface
	responder responder.Interface
}

func New(args Args) *Command {
	server := Command{
		interactionFlow:  args.InteractionFlow,
		responder: args.ResponderService,
		discordClient:    args.DiscordClient,
		serverService:    args.ServerService,
	}

	args.DiscordClient.Router.Command(
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

func (command *Command) execute(ctx *bot.InteractionContext) {
	command.ServerSetup(ctx, command.discordClient.Router)
}

func (command *Command) ServerSetup(ctx *bot.InteractionContext, router *bot.Router) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	menuID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitStringSelect(
		ctx,
		[]components.SelectOption{
			components.NewSelectOptionBuilder().
				SetLabel("Obter configurações").
				SetValue("obtain").
				SetDescription("Consultar as configurações atuais do servidor").
				Build(),
			components.NewSelectOptionBuilder().
				SetLabel("Criar/Atualizar configurações").
				SetValue("upsert").
				SetDescription("Criar ou atualizar as configurações do servidor").
				Build(),
		}, components.NewStringSelectMenuBuilder().
			SetCustomID(menuID).
			SetPlaceholder("Selecione uma opção").
			SetMinValues(1).
			SetMaxValues(1).
			Build(),
	)
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	switch interaction.Values()[0] {
	case "obtain":
		command.Obtain(interaction)
		return
	case "upsert":
		command.Create(interaction)
		return
	default:
		return
	}
}

func (command *Command) Obtain(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	data, err := command.serverService.ObtainServerWithConfig(ctx.Context(), ctx.Guild.ID.String())
	if err != nil {
		if errors.Is(err, server.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	if data.Config == nil {
		command.responder.Sender(responder.WithCustomMessage(ctx, "config does not exist", "", nil, true))
		return
	}

	command.responder.Sender(responder.WithStatusOk(ctx, nil))
}

func (command *Command) Create(ctx *bot.InteractionContext) {
	inputID := uuid.NewString()
	modalID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitSimpleTextModal(ctx,
		components.NewTextInputBuilder().
			SetCustomID(inputID).
			SetLabel("Digite um prefixo para o bot").
			SetStyle(components.TextInputStyleShort).
			SetPlaceholder("e!").
			SetRequired(true).
			SetMaxLength(1).
			Build(),
		components.NewModalBuilder().SetCustomID(modalID).SetTitle("Adicionar/Atualizar as configurações").Build(),
	)
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	err = command.serverService.CreateServerWithConfig(interaction.Context(), domain.Server{
		GuildID: interaction.Guild.ID.String(),
		Config: &domain.Config{
			GuildID: interaction.Guild.ID.String(),
			Prefix:  interaction.ModalValue(inputID),
		},
	})
	if err != nil {
		if errors.Is(err, server.ErrConflict) {
			command.responder.Sender(responder.WithStatusConflict(ctx, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	command.responder.Sender(responder.WithStatusCreated(interaction, nil))
}
