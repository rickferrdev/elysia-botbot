package autoreactions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/rest"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
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
		client    *discord.DiscordClient
		logger    *slog.Logger
	}

	Params struct {
		fx.In
		Responder              responder.Interface
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
		responder: params.Responder,
		client:    params.DiscordClient,
		logger:    params.Logger,
		database:  params.AutoReactionsInterface,
	}

	return &service
}

type GetChannelID struct {
	channelID   string
	interaction *bot.InteractionContext
}

type GetEmoji struct {
	emojiName   string
	interaction *bot.InteractionContext
}

type GetType struct {
	attachmentType domain.AutoReactionType
	interaction    *bot.InteractionContext
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

	options := []components.SelectOption{
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
	}

	menuID := uuid.NewString()
	selectMenuBuilder := components.NewStringSelectMenuBuilder().
		SetCustomID(menuID).
		SetMaxValues(1).
		SetMinValues(1).
		AddOptions(options[0], options[1], options[2], options[3]).
		SetPlaceholder("Selecione uma opção").
		Build()

	row := components.NewActionRowBuilder().AddComponents(selectMenuBuilder)
	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		service.responder.SendError(ctx, err)
		return
	}
	timeout, cancel := context.WithTimeout(ctx.Context(), 2*time.Minute)
	defer cancel()
	interaction, err := ctx.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == menuID && ic.User().ID == ctx.User().ID
	})
	if err != nil {
		service.responder.SendError(ctx, err)
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
	channelSelect := components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		SetPlaceholder("Selecione um canal").
		Build()

	row := components.NewActionRowBuilder().AddComponents(channelSelect).Build()
	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		return err
	}

	timeout, cancel := context.WithTimeout(ctx.Context(), 2*time.Minute)
	defer cancel()
	interaction, err := ctx.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == channelSelectID &&
			ic.User().ID == ctx.User().ID
	})
	if err != nil {
		return err
	}

	channelID := interaction.Values()[0]

	menuID, inputID := uuid.NewString(), uuid.NewString()
	input := components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetPlaceholder("Qual emoji usar?").
		SetStyle(components.TextInputStyleShort).
		SetRequired(true).
		SetLabel("Emoji").
		Build()

	row = components.NewActionRowBuilder().AddComponents(input).Build()
	modal := components.NewModalBuilder().
		SetTitle("Configurações").
		SetCustomID(menuID).
		AddComponents(row).
		Build()

	if err := interaction.ShowModalBuilder(modal); err != nil {
		return err
	}

	timeout, cancel = context.WithTimeout(interaction.Context(), 2*time.Minute)
	defer cancel()
	interaction, err = interaction.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.IsModalSubmit() &&
			ic.CustomID() == menuID &&
			ic.User().ID == ctx.User().ID
	})
	if err != nil {
		return err
	}

	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	emoji := interaction.ModalValue(inputID)

	options := []components.SelectOption{
		components.NewSelectOptionBuilder().SetLabel("Mensagem").SetValue(string(domain.AutoReactionTypeMessage)).SetDescription("Monitorar as mensagens recebidas no canal").Build(),
		components.NewSelectOptionBuilder().SetLabel("Anexos").SetValue(string(domain.AutoReactionTypeAttachment)).SetDescription("Monitorar os anexos recebidos no canal").Build(),
	}

	stringSelectMenuID := uuid.NewString()
	stringSelectMenu := components.NewStringSelectMenuBuilder().
		SetCustomID(stringSelectMenuID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione uma opção").
		AddOptions(options[0], options[1]).
		Build()

	row = components.NewActionRowBuilder().AddComponents(stringSelectMenu).Build()
	if _, err := interaction.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		return err
	}

	timeout, cancel = context.WithTimeout(ctx.Context(), 2*time.Minute)
	defer cancel()
	interaction, err = ctx.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == stringSelectMenuID &&
			ic.User().ID == ctx.User().ID
	})
	if err != nil {
		return err
	}

	if err := interaction.DeferUpdate(); err != nil {
		return err
	}

	attachmentType := interaction.Values()[0]

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
	channelSelectMenu := components.NewChannelSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione o canal monitorado").
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		Build()

	row := components.NewActionRowBuilder().AddComponents(channelSelectMenu).Build()
	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		return err
	}

	timeout, cancel := context.WithTimeout(ctx.Context(), 2*time.Minute)
	defer cancel()
	interaction, err := ctx.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == menuID &&
			ic.User().ID == ctx.User().ID
	})
	if err != nil {
		return err
	}

	if err := interaction.DeferUpdate(); err != nil {
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

	menuID := uuid.NewString()
	channelSelectMenu := components.NewChannelSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione o canal monitorado").
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		Build()

	row := components.NewActionRowBuilder().AddComponents(channelSelectMenu).Build()
	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		return err
	}

	timeout, cancel := context.WithTimeout(ctx.Context(), 2*time.Minute)
	defer cancel()
	interaction, err := ctx.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == menuID &&
			ic.User().ID == ctx.User().ID
	})
	if err != nil {
		return err
	}

	channelID := interaction.Values()[0]

	menuID, inputID := uuid.NewString(), uuid.NewString()
	input := components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetPlaceholder("Qual emoji usar?").
		SetStyle(components.TextInputStyleShort).
		SetRequired(true).
		SetLabel("Emoji").
		Build()

	row = components.NewActionRowBuilder().AddComponents(input).Build()
	modal := components.NewModalBuilder().
		SetTitle("Configurações").
		SetCustomID(menuID).
		AddComponents(row).
		Build()

	if err := interaction.ShowModalBuilder(modal); err != nil {
		return err
	}

	timeout, cancel = context.WithTimeout(interaction.Context(), 2*time.Minute)
	defer cancel()
	interaction, err = interaction.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.IsModalSubmit() &&
			ic.CustomID() == menuID &&
			ic.User().ID == ctx.User().ID
	})
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

	menuID := uuid.NewString()
	channelSelectMenu := components.NewChannelSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione o canal monitorado").
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		Build()

	row := components.NewActionRowBuilder().AddComponents(channelSelectMenu).Build()
	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		return err
	}

	timeout, cancel := context.WithTimeout(ctx.Context(), 2*time.Minute)
	defer cancel()
	interaction, err := ctx.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == menuID &&
			ic.User().ID == ctx.User().ID
	})
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
