package autoreactions

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/components"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discord"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/autoreactions"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		responder responder.Interface
		database  autoreactions.Interface
		flow      interactionflow.Interface
		client    *discord.DiscordClient
		logger    *slog.Logger
	}

	Params struct {
		fx.In
		Responder              responder.Interface
		InteractionFlow        interactionflow.Interface
		AutoReactionsInterface autoreactions.Interface
		Logger                 *slog.Logger
		DiscordClient          *discord.DiscordClient
	}
)

type Interface interface {
	AutoReactionsSetup(ctx *bot.InteractionContext)
	AutoReactionsMessageCreate(ctx *bot.MessageContext)
}

func New(params Params) *Service {
	service := Service{
		flow:      params.InteractionFlow,
		responder: params.Responder,
		client:    params.DiscordClient,
		logger:    params.Logger,
		database:  params.AutoReactionsInterface,
	}

	return &service
}

func (service *Service) AutoReactionsMessageCreate(ctx *bot.MessageContext) {
	if ctx.Author.Bot {
		return
	}

	data, err := service.database.GetOne(ctx.Context(), ctx.GuildID.String(), ctx.ChannelID().String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		service.logger.Error(err.Error())
		return
	}

	switch data.Type {
	case domain.AutoReactionTypeAttachment:
		if len(ctx.Attachments) == 0 {
			return
		}
		if err := ctx.React(data.Emoji); err != nil {
			service.logger.Error(err.Error())
		}
		return
	case domain.AutoReactionTypeMessage:
		if len(ctx.Content) == 0 {
			return
		}
		if err := ctx.React(data.Emoji); err != nil {
			service.logger.Error(err.Error())
		}
		return
	default:
		return
	}
}

func (service *Service) AutoReactionsSetup(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	menuID := uuid.NewString()
	interaction, err := service.flow.AwaitStringSelect(ctx, []components.SelectOption{
		components.NewSelectOptionBuilder().
			SetLabel("Criar").
			SetValue("create").
			SetDescription("Criar um novo registro de mensagem").
			Build(),

		components.NewSelectOptionBuilder().
			SetLabel("Obter").
			SetValue("read").
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
		service.responder.SendError(interaction, err)
		return
	}

	switch interaction.Values()[0] {
	case "create":
		if err := service.Create(interaction); err != nil {
			service.responder.SendError(interaction, err)
			return
		}
		return
	case "read":
		if err := service.Obtain(interaction); err != nil {
			service.responder.SendError(interaction, err)
			return
		}
		return
	case "update":
		if err := service.Update(interaction); err != nil {
			service.responder.SendError(interaction, err)
			return
		}
		return
	case "delete":
		if err := service.Delete(interaction); err != nil {
			service.responder.SendError(interaction, err)
			return
		}
		return
	}
}

func (service *Service) Create(ctx *bot.InteractionContext) error {
	if err := ctx.DeferUpdate(); err != nil {
		return err
	}

	channelSelectID := uuid.NewString()
	interaction, err := service.flow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		SetPlaceholder("Selecione um canal.").
		Build())
	if err != nil {
		return err
	}

	channelID := interaction.Values()[0]
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err = service.flow.AwaitSimpleTextModal(interaction, components.NewTextInputBuilder().
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
		return err
	}

	emoji := interaction.ModalValue(inputID)
	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	stringSelectMenuID := uuid.NewString()
	interaction, err = service.flow.AwaitStringSelect(interaction, []components.SelectOption{
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
		return err
	}

	attachmentType := interaction.Values()[0]
	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	autoreactions := domain.AutoReactions{
		GuildID:   interaction.Guild.ID.String(),
		ChannelID: channelID,
		Type:      domain.AutoReactionType(attachmentType),
		Emoji:     emoji,
	}

	res, err := service.database.CreateOne(interaction.Context(), autoreactions)
	if err != nil {
		return err
	}

	service.responder.SendMessage(interaction, responder.Message{
		Title:            "Tudo criado",
		Description:      fmt.Sprintf("**Identificador:** %d\n", res.ID),
		WithTimestampNow: true,
		WithColor:        true,
		ClearComponents:  true,
	})

	return nil
}

func (service *Service) Obtain(ctx *bot.InteractionContext) error {
	if err := ctx.DeferUpdate(); err != nil {
		return err
	}

	menuID := uuid.NewString()
	interaction, err := service.flow.AwaitChannelSelect(ctx, components.NewChannelSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione o canal monitorado").
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		Build())
	if err != nil {
		return err
	}

	channelID := interaction.Values()[0]

	data, err := service.database.GetOne(ctx.Context(), interaction.Guild.ID.String(), channelID)
	if err != nil {
		return err
	}

	service.responder.SendMessage(interaction, responder.Message{
		Title:            fmt.Sprintf("**Identificador:** %d", data.ID),
		WithTimestampNow: true,
		WithColor:        true,
		ClearComponents:  true,
	})

	return nil
}

func (service *Service) Update(ctx *bot.InteractionContext) error {
	if err := ctx.DeferUpdate(); err != nil {
		return err
	}

	channelSelectID := uuid.NewString()
	interaction, err := service.flow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione um canal.").
		Build())
	if err != nil {
		return err
	}

	channelID := interaction.Values()[0]

	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err = service.flow.AwaitSimpleTextModal(interaction, components.NewTextInputBuilder().
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
		return err
	}

	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	emoji := interaction.ModalValue(inputID)
	data, err := service.database.UpdateOne(ctx.Context(), interaction.Guild.ID.String(), channelID, emoji)
	if err != nil {
		return err
	}

	service.responder.SendMessage(interaction, responder.Message{
		Title:            fmt.Sprintf("**Identificador:** %d", data.ID),
		WithTimestampNow: true,
		WithColor:        true,
		ClearComponents:  true,
	})

	return nil
}

func (service *Service) Delete(ctx *bot.InteractionContext) error {
	if err := ctx.DeferUpdate(); err != nil {
		return err
	}

	channelSelectID := uuid.NewString()
	interaction, err := service.flow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione um canal.").
		Build())
	if err != nil {
		return err
	}

	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	channelID := interaction.Values()[0]
	data, err := service.database.DeleteOne(ctx.Context(), interaction.Guild.ID.String(), channelID)
	if err != nil {
		return err
	}

	service.responder.SendMessage(interaction, responder.Message{
		Title:            fmt.Sprintf("**Identificador:** %d", data.ID),
		WithTimestampNow: true,
		WithColor:        true,
		ClearComponents:  true,
	})

	return nil
}
