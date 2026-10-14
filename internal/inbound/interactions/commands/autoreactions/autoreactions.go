package autoreactions

import (
	"errors"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/permissions"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/autoreactions"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type Args struct {
	fx.In
	ResponderService     responder.Interface
	InteractionFlow      interactionflow.Interface
	AutoReactionsService autoreactions.Interface
	DiscordClient        *discord.DiscordClient
}

type Command struct {
	responderService     responder.Interface
	interactionFlow      interactionflow.Interface
	autoReactionsService autoreactions.Interface
	discordClient        *discord.DiscordClient
}

func New(args Args) *Command {
	server := Command{
		responderService:     args.ResponderService,
		interactionFlow:      args.InteractionFlow,
		discordClient:        args.DiscordClient,
		autoReactionsService: args.AutoReactionsService,
	}

	args.DiscordClient.Router.Command(
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

func (command *Command) execute(ctx *bot.InteractionContext) {
	command.AutoReactionsSetup(ctx)
}

func (command *Command) AutoReactionsSetup(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	menuID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitStringSelect(ctx, []components.SelectOption{
		components.NewSelectOptionBuilder().
			SetLabel("Criar").
			SetValue("create").
			SetDescription("Criar um novo registro de mensagem").
			Build(),

		components.NewSelectOptionBuilder().
			SetLabel("Obter").
			SetValue("obtain").
			SetDescription("Consultar mensagens e seus anexos").
			Build(),

		components.NewSelectOptionBuilder().
			SetLabel("Atualizar").
			SetValue("update").
			SetDescription("Atualizar um registro existente").
			Build(),

		components.NewSelectOptionBuilder().
			SetLabel("Deletar").
			SetValue("delete").
			SetDescription("Remover um registro existente").
			Build(),
	}, components.NewStringSelectMenuBuilder().
		SetCustomID(menuID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione uma opção").
		Build(),
	)
	if err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	switch interaction.Values()[0] {
	case "create":
		command.Create(interaction)
		return
	case "obtain":
		command.Obtain(interaction)
		return
	case "update":
		command.Update(interaction)
		return
	case "delete":
		command.Delete(interaction)
		return
	default:
		return
	}
}

func (command *Command) Create(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	channelSelectID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		SetPlaceholder("Selecione um canal.").
		Build())
	if err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	channelID := interaction.Values()[0]
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err = command.interactionFlow.AwaitSimpleTextModal(interaction, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetPlaceholder("Qual emoji usar?").
		SetStyle(components.TextInputStyleShort).
		SetRequired(true).
		SetLabel("Emoji").
		Build(), components.NewModalBuilder().
		SetTitle("Configurações").
		SetCustomID(menuID).
		Build())
	if err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	emoji := interaction.ModalValue(inputID)
	if err := interaction.DeferUpdate(); err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	stringSelectMenuID := uuid.NewString()
	interaction, err = command.interactionFlow.AwaitStringSelect(interaction, []components.SelectOption{
		components.NewSelectOptionBuilder().
			SetLabel("Mensagem").
			SetValue(string(domain.AutoReactionTypeMessage)).
			SetDescription("Monitorar as mensagens recebidas no canal").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Anexos").
			SetValue(string(domain.AutoReactionTypeAttachment)).
			SetDescription("Monitorar os anexos recebidos no canal").
			Build(),
	}, components.NewStringSelectMenuBuilder().
		SetCustomID(stringSelectMenuID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione uma opção").
		Build())
	if err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	attachmentType := interaction.Values()[0]
	if err := interaction.DeferUpdate(); err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	data := domain.AutoReactions{
		GuildID:   interaction.Guild.ID.String(),
		ChannelID: channelID,
		Type:      domain.AutoReactionType(attachmentType),
		Emoji:     emoji,
	}

	_, err = command.autoReactionsService.CreateOne(interaction.Context(), data)
	if err != nil {
		if errors.Is(err, autoreactions.ErrConflict) {
			command.responderService.Sender(responder.WithStatusConflict(ctx, err))
			return
		}
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	command.responderService.Sender(responder.WithStatusCreated(interaction, nil))
}

func (command *Command) Obtain(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	menuID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitChannelSelect(ctx, components.NewChannelSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione o canal monitorado").
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		Build())
	if err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	channelID := interaction.Values()[0]

	_, err = command.autoReactionsService.ObtainOne(ctx.Context(), interaction.Guild.ID.String(), channelID)
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			command.responderService.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	command.responderService.Sender(responder.WithStatusOk(interaction, nil))
}

func (command *Command) Update(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	channelSelectID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione um canal.").
		Build())
	if err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	channelID := interaction.Values()[0]
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err = command.interactionFlow.AwaitSimpleTextModal(interaction, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetPlaceholder("Qual emoji usar?").
		SetStyle(components.TextInputStyleShort).
		SetRequired(true).
		SetLabel("Emoji").
		Build(), components.NewModalBuilder().
		SetTitle("Configurações").
		SetCustomID(menuID).
		Build())
	if err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	emoji := interaction.ModalValue(inputID)
	_, err = command.autoReactionsService.UpdateOne(ctx.Context(), interaction.Guild.ID.String(), channelID, emoji)
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			command.responderService.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	command.responderService.Sender(responder.WithStatusOk(interaction, nil))
}

func (command *Command) Delete(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	channelSelectID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione um canal.").
		Build())
	if err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	channelID := interaction.Values()[0]
	_, err = command.autoReactionsService.DeleteOne(ctx.Context(), interaction.Guild.ID.String(), channelID)
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			command.responderService.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responderService.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	command.responderService.Sender(responder.WithStatusOk(interaction, nil))
}
