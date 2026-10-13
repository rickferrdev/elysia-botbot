package embeds

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/embeds"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type MenuID struct {
	interaction *bot.InteractionContext
	id          int
}

type (
	Service struct {
		flow      interactionflow.Interface
		database  embeds.Interface
		responder responder.Interface
	}

	Params struct {
		fx.In
		InteractionFlow interactionflow.Interface
		EmbedsDatabase  embeds.Interface
		Responder       responder.Interface
	}
)

type Interface interface {
	EmbedsSetup(ctx *bot.InteractionContext)
}

func New(params Params) *Service {
	service := Service{
		flow:      params.InteractionFlow,
		responder: params.Responder,
		database:  params.EmbedsDatabase,
	}

	return &service
}

func (service *Service) EmbedsSetup(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	menuID := uuid.NewString()
	interaction, err := service.flow.AwaitStringSelect(ctx, []components.SelectOption{
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
		service.responder.Sender(
			responder.WithStatusInternalServerError(ctx, err),
		)
		return
	}

	switch interaction.Values()[0] {
	case "create":
		if err := service.Create(interaction); err != nil {
			service.responder.Sender(
				responder.WithStatusInternalServerError(ctx, err),
			)
		}
		return
	case "delete":
		if err := service.Delete(interaction); err != nil {
			service.responder.Sender(
				responder.WithStatusInternalServerError(ctx, err),
			)
		}
		return
	case "obtain":
		if err := service.Obtain(interaction); err != nil {
			service.responder.Sender(
				responder.WithStatusInternalServerError(ctx, err),
			)
		}
		return
	case "publish":
		if err := service.Publish(interaction); err != nil {
			service.responder.Sender(
				responder.WithStatusInternalServerError(ctx, err),
			)
		}
		return
	case "update":
		if err := service.Update(interaction); err != nil {
			service.responder.Sender(
				responder.WithStatusInternalServerError(ctx, err),
			)
		}
		return
	default:
		return
	}
}

func (service *Service) Create(ctx *bot.InteractionContext) error {
	if err := ctx.DeferUpdate(); err != nil {
		return err
	}

	content := "Envie o JSON como arquivo ou cole o conteúdo JSON nesta conversa."
	userID, channelID := ctx.User().ID, ctx.ChannelID()
	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Content:    &content,
		Components: &[]components.Component{},
	}); err != nil {
		return err
	}

	await, cancel := context.WithTimeout(ctx.Context(), 2*time.Minute)
	defer cancel()
	message, err := ctx.Bot.AwaitMessage(await, func(mc *bot.MessageContext) bool {
		return mc.Author.ID == userID &&
			mc.ChannelID() == channelID &&
			(len(mc.Attachments) > 0 || strings.TrimSpace(mc.Content) != "")
	})
	if err != nil {
		return err
	}

	switch {
	case len(message.Attachments) > 0:
		if err := service.HandlerAttachments(message); err != nil {
			if core.IsCode(err, core.CodeBadRequest) {
				service.responder.Sender(responder.WithStatusBadRequest(ctx, err))
				return nil
			}
			return err
		}
	case len(message.Content) > 0:
		if err := service.HandlerMessageContent(message, 0); err != nil {
			if core.IsCode(err, core.CodeBadRequest) {
				service.responder.Sender(responder.WithStatusBadRequest(ctx, err))
				return nil
			}
			return err
		}
	default:
		return core.NewAppErrorWithStatus(
			"message must contain JSON content or an attachment",
			core.CodeBadRequest,
			core.StatusBadRequest,
			nil,
		)
	}

	return service.responder.Sender(responder.WithStatusCreated(ctx, nil))
}

func (service *Service) Obtain(ctx *bot.InteractionContext) error {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := service.flow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Digite o identificador").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Busca").SetCustomID(menuID).Build(),
	)
	if err != nil {
		return err
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	embed, err := service.database.GetOne(interaction.Context(), id)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			service.responder.Sender(responder.WithStatusNotFound(interaction, err))
			return nil
		}
		return err
	}

	var decode domain.EmbedContentDecode
	if err := json.Unmarshal([]byte(embed.Content), &decode); err != nil {
		return err
	}

	if len(decode.Content) > 0 {
		if _, err := interaction.EditReplyComplex(rest.EditMessageParams{
			Content: &decode.Content,
			Embeds:  &decode.Embeds,
		}); err != nil {
			return err
		}
	} else {
		return service.responder.Sender(responder.WithStatusOk(interaction, nil))
	}

	return nil
}

func (service *Service) Publish(ctx *bot.InteractionContext) error {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := service.flow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Digite o identificador").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Busca").SetCustomID(menuID).Build(),
	)
	if err != nil {
		return err
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	channelMenuID := uuid.NewString()
	interaction, err = service.flow.AwaitChannelSelect(interaction,
		components.NewChannelSelectMenuBuilder().
			SetCustomID(channelMenuID).
			SetMaxValues(1).
			SetMinValues(1).
			SetPlaceholder("Selecione qual canal deseja enviar").
			SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).Build())
	if err != nil {
		return err
	}

	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	channelID, err := snowflake.Parse(interaction.Values()[0])
	if err != nil {
		return err
	}

	data, err := service.database.GetOne(interaction.Context(), id)
	if err != nil {
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			service.responder.Sender(responder.WithStatusNotFound(interaction, err))
			return nil
		}
		return err
	}

	var decode domain.EmbedContentDecode
	if err := json.Unmarshal([]byte(data.Content), &decode); err != nil {
		return err
	}

	if _, err := interaction.Bot.Rest.CreateMessageComplex(interaction.Context(), channelID, messages.MessageSend{
		Content: decode.Content,
		Embeds:  decode.Embeds,
	}); err != nil {
		return err
	}

	return service.responder.Sender(responder.WithStatusCreated(interaction, nil))
}

func (service *Service) Update(ctx *bot.InteractionContext) error {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := service.flow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Digite o identificador").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Busca").SetCustomID(menuID).Build(),
	)
	if err != nil {
		return err
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	if _, err := interaction.EditReply("Envie o JSON como arquivo ou cole o conteúdo JSON nesta conversa."); err != nil {
		return err
	}

	await, cancel := context.WithTimeout(interaction.Context(), 2*time.Minute)
	defer cancel()
	message, err := ctx.Bot.AwaitMessage(await, func(mc *bot.MessageContext) bool {
		return mc.Author.ID == interaction.User().ID &&
			mc.ChannelID() == interaction.ChannelID() &&
			(len(mc.Attachments) > 0 || strings.TrimSpace(mc.Content) != "")
	})
	if err != nil {
		return err
	}

	switch {
	case len(message.Attachments) > 0:
		if err := service.HandlerAttachments(message); err != nil {
			if core.IsCode(err, core.CodeBadRequest) {
				service.responder.Sender(responder.WithStatusBadRequest(interaction, err))
				return nil
			}
			return err
		}
	case len(message.Content) > 0:
		if err := service.HandlerMessageContent(message, id); err != nil {
			if core.IsCode(err, core.CodeBadRequest) {
				service.responder.Sender(responder.WithStatusBadRequest(interaction, err))
				return nil
			}
			return err
		}
	default:
		return core.NewAppErrorWithStatus(
			"message must contain JSON content or an attachment",
			core.CodeBadRequest,
			core.StatusBadRequest,
			nil,
		)
	}

	return service.responder.Sender(responder.WithStatusOk(interaction, nil))
}

func (service *Service) Delete(ctx *bot.InteractionContext) error {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err := service.flow.AwaitSimpleTextModal(ctx, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Digite o identificador").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build(),
		components.NewModalBuilder().SetTitle("Busca").SetCustomID(menuID).Build(),
	)
	if err != nil {
		return err
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	_, err = service.database.DeleteOne(interaction.Context(), id)
	if err != nil {
		return err
	}

	return service.responder.Sender(responder.WithStatusNoContent(interaction, nil))
}

func (service *Service) HandlerAttachments(message *bot.MessageContext) error {
	return errors.New("embed attachment handling is not implemented")
}

func (service *Service) HandlerMessageContent(message *bot.MessageContext, id int) error {
	formatted := strings.TrimSpace(message.Content)
	if !json.Valid([]byte(formatted)) {
		return core.NewAppErrorWithStatus(
			"invalid JSON",
			core.CodeBadRequest,
			core.StatusBadRequest,
			errors.New("invalid json"),
		)
	}

	var embedContent domain.EmbedContentDecode
	if err := json.Unmarshal([]byte(formatted), &embedContent); err != nil {
		return core.NewAppErrorWithStatus(
			"invalid JSON",
			core.CodeBadRequest,
			core.StatusBadRequest,
			err,
		)
	}
	bytes, err := json.Marshal(embedContent)
	if err != nil {
		return err
	}

	embed := domain.Embed{
		ID:        id,
		GuildID:   message.GuildID.String(),
		UserID:    message.Author.ID.String(),
		Content:   string(bytes),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if _, err := service.database.UpsertOne(message.Context(), embed); err != nil {
		return err
	}

	return nil
}
