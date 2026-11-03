package templates

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
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	templateService "github.com/rickferrdev/elysia-botbot/internal/core/services/templates"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/resembeds"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type FxArgs struct {
	fx.In

	Responder       responder.Interface
	InteractionFlow interactionflow.Interface
	Templates       templateService.Interface
	DiscordClient   *discordclient.DiscordClient
}

type Command struct {
	flow      interactionflow.Interface
	templates templateService.Interface
	client    *discordclient.DiscordClient
	responder responder.Interface
}

func New(args FxArgs) *Command {
	command := Command{
		flow:      args.InteractionFlow,
		responder: args.Responder,
		client:    args.DiscordClient,
		templates: args.Templates,
	}

	args.DiscordClient.Router.Command(
		"templates",
		"Crie, consulte, publique e atualize seus templates.",
		command.execute,
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

	return &command
}

func decodeTemplateContent(template *domain.Template) (domain.EmbedContentDecode, error) {
	if template.TemplateContent == nil {
		return domain.EmbedContentDecode{}, errors.New("template content is missing")
	}

	decoded := domain.EmbedContentDecode{
		Content: template.TemplateContent.Content,
	}
	if template.TemplateContent.Embeds != "" {
		if err := json.Unmarshal([]byte(template.TemplateContent.Embeds), &decoded.Embeds); err != nil {
			return domain.EmbedContentDecode{}, err
		}
	}
	if template.TemplateContent.Attachments != "" {
		if err := json.Unmarshal([]byte(template.TemplateContent.Attachments), &decoded.Attachments); err != nil {
			return domain.EmbedContentDecode{}, err
		}
	}

	return decoded, nil
}

func (cmd *Command) execute(ctx *bot.InteractionContext) {
	cmd.TemplatesSetup(ctx)
}

func (cmd *Command) TemplatesSetup(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	menuID := uuid.NewString()
	interaction, err := cmd.flow.AwaitStringSelect(ctx, []components.SelectOption{
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
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	switch interaction.Values()[0] {
	case "create":
		cmd.Create(interaction)
		return
	case "obtain":
		cmd.Obtain(interaction)
		return
	case "publish":
		cmd.Publish(interaction)
		return
	case "update":
		cmd.Update(interaction)
		return
	case "delete":
		cmd.Delete(interaction)
		return
	case "obtain-all":
		cmd.ObtainAll(interaction)
		return
	default:
		return
	}
}

func (cmd *Command) Create(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	avatar := ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})
	content := "Envie um anexo (ex.: `.json`) ou cole a estrutura JSON da incorporação nesta conversa."
	userID, channelID := ctx.User().ID, ctx.ChannelID()
	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Content:    &content,
		Components: &[]components.Component{},
	}); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
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
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	var source string
	var data *domain.Template

	switch {
	case len(message.Attachments) > 0:
		source = "Anexo"
		data, err = cmd.templates.HandlerAttachments(ctx.Context(), 0, message.GuildID.String(), ctx.User().ID.String(), message.Attachments[0])
		if err != nil {
			if errors.Is(err, templateService.ErrInvalidInput) || errors.Is(err, templateService.ErrMaxSize) {
				cmd.responder.Sender(ctx, responder.WithBadArguments, err)
				return
			}
			if errors.Is(err, templateService.ErrConflict) {
				cmd.responder.Sender(ctx, responder.WithConflict, err)
				return
			}
			cmd.responder.Sender(ctx, responder.WithInternal, err)
			return
		}

	case len(message.Content) > 0:
		source = "Texto (JSON)"
		data, err = cmd.templates.HandlerMessageContent(ctx.Context(), 0, message.GuildID.String(), ctx.User().ID.String(), message.Content)
		if err != nil {
			if errors.Is(err, templateService.ErrInvalidInput) {
				cmd.responder.Sender(ctx, responder.WithBadArguments, err)
				return
			}
			if errors.Is(err, templateService.ErrConflict) {
				cmd.responder.Sender(ctx, responder.WithConflict, err)
				return
			}
			cmd.responder.Sender(ctx, responder.WithInternal, err)
			return
		}

	default:
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	embed := resembeds.EmbedCreateEmbeds(avatar, *data, source)
	cmd.responder.Sender(ctx, responder.WithEmbed(embed), nil)
}

func (cmd *Command) Obtain(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	avatar := ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})
	interaction, err := cmd.flow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
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
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("%w: invalid template ID: %v", templateService.ErrInvalidInput, err)
		cmd.responder.Sender(interaction, responder.WithBadArguments, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	embed, err := cmd.templates.ObtainOne(interaction.Context(), id)
	if err != nil {
		if errors.Is(err, templateService.ErrNotFound) {
			cmd.responder.Sender(interaction, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	decode, err := decodeTemplateContent(embed)
	if err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	cmd.responder.Sender(interaction, responder.WithEmbed(
		resembeds.EmbedObtainEmbeds(avatar, *embed, decode),
	), nil)
}

func (cmd *Command) ObtainAll(ctx *bot.InteractionContext) {
	avatar := ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	data, err := cmd.templates.ObtainAll(ctx.Context(), ctx.GuildID().String(), ctx.User().ID.String())
	if err != nil {
		if errors.Is(err, templateService.ErrNotFound) {
			cmd.responder.Sender(ctx, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	cmd.responder.Sender(ctx, responder.WithEmbed(
		resembeds.EmbedObtainAllEmbeds(avatar, data),
	), nil)
}

func (cmd *Command) Publish(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	avatar := ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})
	interaction, err := cmd.flow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
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
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("%w: invalid template ID: %v", templateService.ErrInvalidInput, err)
		cmd.responder.Sender(interaction, responder.WithBadArguments, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	channelMenuID := uuid.NewString()
	interaction, err = cmd.flow.AwaitChannelSelect(interaction,
		components.NewChannelSelectMenuBuilder().
			SetCustomID(channelMenuID).
			SetMaxValues(1).
			SetMinValues(1).
			SetPlaceholder("Selecione o canal para publicar.").
			SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).Build())
	if err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	channelID, err := snowflake.Parse(interaction.Values()[0])
	if err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	data, err := cmd.templates.ObtainOne(interaction.Context(), id)
	if err != nil {
		if errors.Is(err, templateService.ErrNotFound) {
			cmd.responder.Sender(interaction, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	decode, err := decodeTemplateContent(data)
	if err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	if _, err := interaction.Bot.Rest.CreateMessageComplex(interaction.Context(), channelID, messages.MessageSend{
		Content: decode.Content,
		Embeds:  decode.Embeds,
	}); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	cmd.responder.Sender(interaction, responder.WithEmbed(
		resembeds.EmbedPublishEmbeds(avatar, channelID.String(), ctx.User().ID.String(), *data),
	), nil)
}

func (cmd *Command) Update(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	avatar := ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})
	interaction, err := cmd.flow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
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
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("%w: invalid template ID: %v", templateService.ErrInvalidInput, err)
		cmd.responder.Sender(interaction, responder.WithBadArguments, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	content := "Envie o arquivo `.json` ou cole a nova estrutura JSON da incorporação nesta conversa."
	if _, err := interaction.EditReplyComplex(rest.EditMessageParams{
		Content:    &content,
		Components: &[]components.Component{},
	}); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
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
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	var source string
	var data *domain.Template

	switch {
	case len(message.Attachments) > 0:
		source = "Anexo (.json)"
		data, err = cmd.templates.HandlerAttachments(ctx.Context(), id, message.GuildID.String(), ctx.User().ID.String(), message.Attachments[0])
		if err != nil {
			if errors.Is(err, templateService.ErrInvalidInput) || errors.Is(err, templateService.ErrMaxSize) {
				cmd.responder.Sender(interaction, responder.WithBadArguments, err)
				return
			}
			if errors.Is(err, templateService.ErrConflict) {
				cmd.responder.Sender(interaction, responder.WithConflict, err)
				return
			}
			if errors.Is(err, templateService.ErrNotFound) {
				cmd.responder.Sender(interaction, responder.WithNotFound, err)
				return
			}
			cmd.responder.Sender(interaction, responder.WithInternal, err)
			return
		}

	case len(message.Content) > 0:
		source = "Texto (JSON)"
		data, err = cmd.templates.HandlerMessageContent(ctx.Context(), id, message.GuildID.String(), ctx.User().ID.String(), message.Content)
		if err != nil {
			if errors.Is(err, templateService.ErrInvalidInput) {
				cmd.responder.Sender(interaction, responder.WithBadArguments, err)
				return
			}
			if errors.Is(err, templateService.ErrConflict) {
				cmd.responder.Sender(interaction, responder.WithConflict, err)
				return
			}
			if errors.Is(err, templateService.ErrNotFound) {
				cmd.responder.Sender(interaction, responder.WithNotFound, err)
				return
			}
			cmd.responder.Sender(interaction, responder.WithInternal, err)
			return
		}

	default:
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	cmd.responder.Sender(interaction, responder.WithEmbed(
		resembeds.EmbedUpdateEmbeds(avatar, interaction.User().ID.String(), *data, source),
	), nil)
}

func (cmd *Command) Delete(ctx *bot.InteractionContext) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	avatar := ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})
	interaction, err := cmd.flow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
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
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("%w: invalid template ID: %v", templateService.ErrInvalidInput, err)
		cmd.responder.Sender(interaction, responder.WithBadArguments, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	_, err = cmd.templates.DeleteOne(interaction.Context(), id)
	if err != nil {
		if errors.Is(err, templateService.ErrNotFound) {
			cmd.responder.Sender(interaction, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	cmd.responder.Sender(interaction, responder.WithEmbed(
		resembeds.EmbedDeleteEmbeds(avatar, interaction.User().ID.String(), id),
	), nil)
}
