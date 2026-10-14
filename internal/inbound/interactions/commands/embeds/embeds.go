package embeds

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/embeds"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type Args struct {
	fx.In

	ResponderService responder.Interface
	InteractionFlow  interactionflow.Interface
	EmbedsService    embeds.Interface
	DiscordClient    *discord.DiscordClient
}

type Command struct {
	interactionFlow interactionflow.Interface
	embedsService   embeds.Interface
	discordClient   *discord.DiscordClient
	responder       responder.Interface
}

func New(args Args) *Command {
	embeds := Command{
		interactionFlow: args.InteractionFlow,
		responder:       args.ResponderService,
		discordClient:   args.DiscordClient,
		embedsService:   args.EmbedsService,
	}

	args.DiscordClient.Router.Command(
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

func (command *Command) execute(ctx *bot.InteractionContext) {
	command.EmbedsSetup(ctx)
}

func (command *Command) EmbedsSetup(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	menuID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitStringSelect(ctx, []components.SelectOption{
		components.NewSelectOptionBuilder().
			SetLabel("Criar incorporação").
			SetDescription("Cria uma nova incorporação.").
			SetValue("create").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Consultar incorporação").
			SetDescription("Busca e exibe uma incorporação existente.").
			SetValue("obtain").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Publicar incorporação").
			SetDescription("Envia uma incorporação para um canal.").
			SetValue("publish").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Atualizar incorporação").
			SetDescription("Atualiza e exibe uma incorporação existente..").
			SetValue("update").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Excluir incorporação").
			SetDescription("Remove uma incorporação existente.").
			SetValue("delete").
			Build(),
	}, components.NewStringSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione a opção abaixo").
		SetMaxValues(1).
		SetMinValues(1).
		Build(),
	)
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	switch interaction.Values()[0] {
	case "create":
		command.Create(interaction)
		return
	case "delete":
		command.Delete(interaction)
		return
	case "obtain":
		command.Obtain(interaction)
		return
	case "publish":
		command.Publish(interaction)
		return
	case "update":
		command.Update(interaction)
		return
	default:
		return
	}
}

func (command *Command) Create(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	content := "Envie o JSON como arquivo ou cole o conteúdo JSON nesta conversa."
	userID, channelID := ctx.User().ID, ctx.ChannelID()
	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Content:    &content,
		Components: &[]components.Component{},
	}); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	await, cancel := context.WithTimeout(ctx.Context(), 2*time.Minute)
	defer cancel()
	message, err := ctx.Bot.AwaitMessage(await, func(mc *bot.MessageContext) bool {
		return mc.Author.ID == userID &&
			mc.ChannelID() == channelID &&
			(len(mc.Attachments) > 0 || strings.TrimSpace(mc.Content) != "")
	})
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	switch {
	case len(message.Attachments) > 0:
		err = command.embedsService.HandlerAttachments(ctx.Context(), 0, message.GuildID.String(), ctx.User().ID.String(), message.Attachments[0])
		if err != nil {
			if errors.Is(err, embeds.ErrInvalidInput) || errors.Is(err, embeds.ErrMaxSize) {
				command.responder.Sender(responder.WithStatusBadRequest(ctx, err))
				return
			}
			if errors.Is(err, embeds.ErrConflict) {
				command.responder.Sender(responder.WithStatusConflict(ctx, err))
				return
			}
			command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
			return
		}
	case len(message.Content) > 0:
		err = command.embedsService.HandlerMessageContent(ctx.Context(), 0, message.GuildID.String(), ctx.User().ID.String(), message.Content)
		if err != nil {
			if errors.Is(err, embeds.ErrInvalidInput) {
				command.responder.Sender(responder.WithStatusBadRequest(ctx, err))
				return
			}
			if errors.Is(err, embeds.ErrConflict) {
				command.responder.Sender(responder.WithStatusConflict(ctx, err))
				return
			}
			command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
			return
		}
	default:
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	command.responder.Sender(responder.WithStatusCreated(ctx, nil))
}

func (command *Command) Obtain(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := command.interactionFlow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Digite o identificador").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Busca").SetCustomID(menuID).Build(),
	)
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("%w: invalid embed ID: %v", embeds.ErrInvalidInput, err)
		command.responder.Sender(responder.WithStatusBadRequest(interaction, err))
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	embed, err := command.embedsService.ObtainOne(interaction.Context(), id)
	if err != nil {
		if errors.Is(err, embeds.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(ctx, err))
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	var decode domain.EmbedContentDecode
	if err := json.Unmarshal([]byte(embed.Content), &decode); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	if len(decode.Content) > 0 {
		if _, err := interaction.EditReplyComplex(rest.EditMessageParams{
			Content:    &decode.Content,
			Embeds:     &decode.Embeds,
			Components: &[]components.Component{},
		}); err != nil {
			command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
			return
		}
	} else {
		command.responder.Sender(responder.WithStatusOk(interaction, nil))
	}
}

func (command *Command) Publish(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := command.interactionFlow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Digite o identificador").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Busca").SetCustomID(menuID).Build(),
	)
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("%w: invalid embed ID: %v", embeds.ErrInvalidInput, err)
		command.responder.Sender(responder.WithStatusBadRequest(interaction, err))
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	channelMenuID := uuid.NewString()
	interaction, err = command.interactionFlow.AwaitChannelSelect(interaction,
		components.NewChannelSelectMenuBuilder().
			SetCustomID(channelMenuID).
			SetMaxValues(1).
			SetMinValues(1).
			SetPlaceholder("Selecione qual canal deseja enviar").
			SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).Build())
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	channelID, err := snowflake.Parse(interaction.Values()[0])
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	data, err := command.embedsService.ObtainOne(interaction.Context(), id)
	if err != nil {
		if errors.Is(err, embeds.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	var decode domain.EmbedContentDecode
	if err := json.Unmarshal([]byte(data.Content), &decode); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	if _, err := interaction.Bot.Rest.CreateMessageComplex(interaction.Context(), channelID, messages.MessageSend{
		Content: decode.Content,
		Embeds:  decode.Embeds,
	}); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	command.responder.Sender(responder.WithStatusCreated(interaction, nil))
}

func (command *Command) Update(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := command.interactionFlow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Digite o identificador").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Busca").SetCustomID(menuID).Build(),
	)
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("%w: invalid embed ID: %v", embeds.ErrInvalidInput, err)
		command.responder.Sender(responder.WithStatusBadRequest(interaction, err))
		return
	}
	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	content := "Envie o JSON como arquivo ou cole o conteúdo JSON nesta conversa"
	if _, err := interaction.EditReplyComplex(rest.EditMessageParams{
		Content:    &content,
		Components: &[]components.Component{},
	}); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	await, cancel := context.WithTimeout(interaction.Context(), 2*time.Minute)
	defer cancel()
	message, err := ctx.Bot.AwaitMessage(await, func(mc *bot.MessageContext) bool {
		return mc.Author.ID == interaction.User().ID &&
			mc.ChannelID() == interaction.ChannelID() &&
			(len(mc.Attachments) > 0 || strings.TrimSpace(mc.Content) != "")
	})
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	switch {
	case len(message.Attachments) > 0:
		err = command.embedsService.HandlerAttachments(ctx.Context(), id, message.GuildID.String(), ctx.User().ID.String(), message.Attachments[0])
		if err != nil {
			if errors.Is(err, embeds.ErrInvalidInput) || errors.Is(err, embeds.ErrMaxSize) {
				command.responder.Sender(responder.WithStatusBadRequest(ctx, err))
				return
			}
			if errors.Is(err, embeds.ErrConflict) {
				command.responder.Sender(responder.WithStatusConflict(ctx, err))
				return
			}
			command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
			return
		}
	case len(message.Content) > 0:
		err = command.embedsService.HandlerMessageContent(ctx.Context(), id, message.GuildID.String(), ctx.User().ID.String(), message.Content)
		if err != nil {
			if errors.Is(err, embeds.ErrInvalidInput) {
				command.responder.Sender(responder.WithStatusBadRequest(ctx, err))
				return
			}
			if errors.Is(err, embeds.ErrConflict) {
				command.responder.Sender(responder.WithStatusConflict(ctx, err))
				return
			}
			command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
			return
		}
	default:
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
}

func (command *Command) Delete(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := command.interactionFlow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Digite o identificador").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Busca").SetCustomID(menuID).Build(),
	)
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("%w: invalid embed ID: %v", embeds.ErrInvalidInput, err)
		command.responder.Sender(responder.WithStatusBadRequest(interaction, err))
		return
	}
	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	_, err = command.embedsService.DeleteOne(interaction.Context(), id)
	if err != nil {
		if errors.Is(err, embeds.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	command.responder.Sender(responder.WithStatusNoContent(interaction, nil))
}
