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
	"github.com/discord-go/discord.go/users"
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
			SetLabel("Criar").
			SetValue("create").
			SetDescription("Criar um novo registro de incorporação").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Obter").
			SetValue("obtain").
			SetDescription("Consultar uma incorporação existente").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Obter todas").
			SetValue("obtain-all").
			SetDescription("Consultar todas as incorporação existentes").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Publicar").
			SetValue("publish").
			SetDescription("Enviar uma incorporação para um canal").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Atualizar").
			SetValue("update").
			SetDescription("Atualizar uma incorporação existente").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Deletar").
			SetValue("delete").
			SetDescription("Remover uma incorporação existente").
			Build(),
	}, components.NewStringSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione uma opção.").
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
	case "obtain":
		command.Obtain(interaction)
		return
	case "publish":
		command.Publish(interaction)
		return
	case "update":
		command.Update(interaction)
		return
	case "delete":
		command.Delete(interaction)
		return
	case "obtain-all":
		command.ObtainAll(interaction)
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
	content := "Envie um anexo (ex.: `.json`) ou cole a estrutura JSON da incorporação nesta conversa."
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
	var source string
	var data *domain.Embed
	switch {
	case len(message.Attachments) > 0:
		source = "Anexo"
		data, err = command.embedsService.HandlerAttachments(ctx.Context(), 0, message.GuildID.String(), ctx.User().ID.String(), message.Attachments[0])
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
		source = "Texto (JSON)"
		data, err = command.embedsService.HandlerMessageContent(ctx.Context(), 0, message.GuildID.String(), ctx.User().ID.String(), message.Content)
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
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Criação de Incorporação 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()
	fields := map[string]string{
		"── .✦ Identificador:": fmt.Sprintf("`%v`", data.ID),
		"⋆˚꩜｡ Tipo de envio:":  fmt.Sprintf("`%v`", source),
		"⋆˚꩜｡ Autor:":          fmt.Sprintf("<@%v>", ctx.User().ID),
		"⋆˚꩜｡ Servidor:":       fmt.Sprintf("**%v**", ctx.Guild.ID),
		"⋆˚꩜｡ Status:":         "`Criada com sucesso!`",
	}
	for index, value := range fields {
		embeder.AddField(index, value, false)
	}
	command.responder.Sender(responder.WithCustomEmbed(ctx, embeder.Build(), nil))
}
func (command *Command) Obtain(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := command.interactionFlow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Identificador").
		SetPlaceholder("Qual o ID da incorporação? e.g. 1").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Consultar incorporação").SetCustomID(menuID).Build(),
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
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	embed, err := command.embedsService.ObtainOne(interaction.Context(), id)
	if err != nil {
		if errors.Is(err, embeds.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(interaction, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	var decode domain.EmbedContentDecode
	if err := json.Unmarshal([]byte(embed.Content), &decode); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Obter uma Incorporação 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()
	fields := map[string]string{
		"── .✦ Identificador:":  fmt.Sprintf("`%v`", embed.ID),
		"⋆˚꩜｡ Criado por:":      fmt.Sprintf("<@%v>", embed.UserID),
		"⋆˚꩜｡ Servidor:":        fmt.Sprintf("**%v**", embed.GuildID),
		"⋆˚꩜｡ Total de embeds:": fmt.Sprintf("`%d`", len(decode.Embeds)),
	}
	if strings.TrimSpace(decode.Content) != "" {
		contentPreview := decode.Content
		if len(contentPreview) > 80 {
			contentPreview = contentPreview[:77] + "..."
		}
		fields["⋆˚꩜｡ Conteúdo de texto:"] = fmt.Sprintf("`%v`", contentPreview)
	}
	for index, value := range fields {
		embeder.AddField(index, value, false)
	}
	command.responder.Sender(responder.WithCustomEmbed(interaction, embeder.Build(), nil))
}
func (command *Command) ObtainAll(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	data, err := command.embedsService.ObtainAll(ctx.Context(), ctx.GuildID().String(), ctx.User().ID.String())
	if err != nil {
		if errors.Is(err, embeds.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Obter uma Incorporação 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()
	var fields map[string]string
	for _, value := range data {
		var builder strings.Builder
		builder.WriteString("⋆˚꩜｡ Criado por:")
		fmt.Fprintf(&builder, "<@%v>\n", value.UserID)
		builder.WriteString("⋆˚꩜｡ Servidor:")
		fmt.Fprintf(&builder, "**%v**\n", value.GuildID)
		builder.WriteString("⋆˚꩜｡ Total de embeds:")
		fmt.Fprintf(&builder, "`%d`\n", len(data))

		fields = map[string]string{
			fmt.Sprintf("── .✦ Identificador: <%v>:", value.ID): builder.String(),
		}
	}
	for index, value := range fields {
		embeder.AddField(index, value, false)
	}
	command.responder.Sender(responder.WithCustomEmbed(ctx, embeder.Build(), nil))
}
func (command *Command) Publish(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := command.interactionFlow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Identificador").
		SetPlaceholder("Qual o ID da incorporação? e.g. 1").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Publicar incorporação").SetCustomID(menuID).Build(),
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
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	channelMenuID := uuid.NewString()
	interaction, err = command.interactionFlow.AwaitChannelSelect(interaction,
		components.NewChannelSelectMenuBuilder().
			SetCustomID(channelMenuID).
			SetMaxValues(1).
			SetMinValues(1).
			SetPlaceholder("Selecione o canal para publicar.").
			SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).Build())
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	channelID, err := snowflake.Parse(interaction.Values()[0])
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	data, err := command.embedsService.ObtainOne(interaction.Context(), id)
	if err != nil {
		if errors.Is(err, embeds.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(interaction, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	var decode domain.EmbedContentDecode
	if err := json.Unmarshal([]byte(data.Content), &decode); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	if _, err := interaction.Bot.Rest.CreateMessageComplex(interaction.Context(), channelID, messages.MessageSend{
		Content: decode.Content,
		Embeds:  decode.Embeds,
	}); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Publicar uma Incorporação 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()
	fields := map[string]string{
		"── .✦ Identificador:": fmt.Sprintf("`%v`", id),
		"⋆˚꩜｡ Canal de envio:": fmt.Sprintf("<#%v>", channelID),
		"⋆˚꩜｡ Publicado por:":  fmt.Sprintf("<@%v>", interaction.User().ID),
		"⋆˚꩜｡ Status:":         "`Publicada com sucesso!`",
	}
	for index, value := range fields {
		embeder.AddField(index, value, false)
	}
	command.responder.Sender(responder.WithCustomEmbed(interaction, embeder.Build(), nil))
}
func (command *Command) Update(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := command.interactionFlow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Identificador").
		SetPlaceholder("Qual o ID da incorporação? e.g. 1").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Atualizar incorporação").SetCustomID(menuID).Build(),
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
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	content := "Envie o arquivo `.json` ou cole a nova estrutura JSON da incorporação nesta conversa."
	if _, err := interaction.EditReplyComplex(rest.EditMessageParams{
		Content:    &content,
		Components: &[]components.Component{},
	}); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
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
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	var source string
	var data *domain.Embed
	switch {
	case len(message.Attachments) > 0:
		source = "Anexo (.json)"
		data, err = command.embedsService.HandlerAttachments(ctx.Context(), id, message.GuildID.String(), ctx.User().ID.String(), message.Attachments[0])
		if err != nil {
			if errors.Is(err, embeds.ErrInvalidInput) || errors.Is(err, embeds.ErrMaxSize) {
				command.responder.Sender(responder.WithStatusBadRequest(interaction, err))
				return
			}
			if errors.Is(err, embeds.ErrConflict) {
				command.responder.Sender(responder.WithStatusConflict(interaction, err))
				return
			}
			command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
			return
		}
	case len(message.Content) > 0:
		source = "Texto (JSON)"
		data, err = command.embedsService.HandlerMessageContent(ctx.Context(), id, message.GuildID.String(), ctx.User().ID.String(), message.Content)
		if err != nil {
			if errors.Is(err, embeds.ErrInvalidInput) {
				command.responder.Sender(responder.WithStatusBadRequest(interaction, err))
				return
			}
			if errors.Is(err, embeds.ErrConflict) {
				command.responder.Sender(responder.WithStatusConflict(interaction, err))
				return
			}
			command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
			return
		}
	default:
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Atualizar uma Incorporação 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()
	fields := map[string]string{
		"── .✦ Identificador:": fmt.Sprintf("`%v`", data.ID),
		"⋆˚꩜｡ Tipo de envio:":  fmt.Sprintf("`%v`", source),
		"⋆˚꩜｡ Atualizado por:": fmt.Sprintf("<@%v>", interaction.User().ID),
		"⋆˚꩜｡ Status:":         "`Atualizada com sucesso!`",
	}
	for index, value := range fields {
		embeder.AddField(index, value, false)
	}
	command.responder.Sender(responder.WithCustomEmbed(interaction, embeder.Build(), nil))
}

func (command *Command) Delete(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := command.interactionFlow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Identificador").
		SetPlaceholder("Qual o ID da incorporação? e.g. 1").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Deletar incorporação").SetCustomID(menuID).Build(),
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
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	_, err = command.embedsService.DeleteOne(interaction.Context(), id)
	if err != nil {
		if errors.Is(err, embeds.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(interaction, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(interaction, err))
		return
	}
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Deletar uma Incorporação 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()
	fields := map[string]string{
		"── .✦ Identificador:": fmt.Sprintf("`%v`", id),
		"⋆˚꩜｡ Removido por:":   fmt.Sprintf("<@%v>", interaction.User().ID),
		"⋆˚꩜｡ Status:":         "`Removida com sucesso!`",
	}
	for index, value := range fields {
		embeder.AddField(index, value, false)
	}
	command.responder.Sender(responder.WithCustomEmbed(interaction, embeder.Build(), nil))
}
