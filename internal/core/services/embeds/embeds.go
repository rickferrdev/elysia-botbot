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
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
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
		database  embeds.Interface
		responder responder.Interface
	}

	Params struct {
		fx.In
		EmbedsDatabase embeds.Interface
		Responder      responder.Interface
	}
)

type Interface interface {
	EmbedsSetup(ctx *bot.InteractionContext)
}

func New(params Params) *Service {
	service := Service{
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
	options := []components.SelectOption{
		components.NewSelectOptionBuilder().SetLabel("Criar").SetDescription("Criar uma incorporação.").SetValue("add").Build(),
		components.NewSelectOptionBuilder().SetLabel("Obter").SetDescription("Obter uma incorporação.").SetValue("obtain").Build(),
		components.NewSelectOptionBuilder().SetLabel("Enviar").SetDescription("Enviar uma incorporação.").SetValue("publish").Build(),
		components.NewSelectOptionBuilder().SetLabel("Deletar").SetDescription("Deletar uma incorporação.").SetValue("delete").Build(),
	}

	selectMenu := components.NewStringSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione a opção abaixo").
		SetMaxValues(1).
		SetMinValues(1).
		AddOptions(options[0], options[1], options[2], options[3]).
		Build()

	row := components.NewActionRowBuilder().AddComponents(selectMenu).Build()
	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		service.responder.SendError(ctx, err)
		return
	}
	await, cancel := context.WithTimeout(ctx.Context(), 1*time.Minute)
	defer cancel()

	ic, err := ctx.Bot.AwaitInteraction(await, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == menuID && ic.User().ID == ctx.User().ID
	})
	if err != nil {
		service.responder.SendError(ctx, err)
		return
	}

	switch ic.Values()[0] {
	case "add":
		err := service.EmbedsCreate(ic)
		if err != nil {
			service.responder.SendError(ic, err)
		}
		return
	case "delete":
		err := service.EmbedsDelete(ic)
		if err != nil {
			service.responder.SendError(ic, err)
		}
		return
	case "obtain":
		err := service.EmbedsObtain(ic)
		if err != nil {
			service.responder.SendError(ic, err)
		}
		return
	case "publish":
		err := service.EmbedsPublish(ic)
		if err != nil {
			service.responder.SendError(ic, err)
		}
		return
	default:
		return
	}
}

func (service *Service) GetMenuID(ctx *bot.InteractionContext) (*MenuID, error) {
	menuID, inputID := uuid.NewString(), uuid.NewString()
	input := components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Digite o identificador").
		SetMinLength(1).
		SetRequired(true).
		SetStyle(components.TextInputStyleShort).
		Build()

	row := components.NewActionRowBuilder().AddComponents(input).Build()
	if err := ctx.ShowModal(menuID, "Digite o ID", row); err != nil {
		return nil, err
	}

	await, cancel := context.WithTimeout(ctx.Context(), 1*time.Minute)
	defer cancel()

	interaction, err := ctx.Bot.AwaitInteraction(await, func(ic *bot.InteractionContext) bool {
		return ic.IsModalSubmit() && ic.User().ID == ctx.User().ID && ic.CustomID() == menuID
	})
	if err != nil {
		return nil, err
	}

	value := interaction.ModalValue(inputID)
	id, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	return &MenuID{interaction, id}, nil
}

func (service *Service) EmbedsCreate(ctx *bot.InteractionContext) error {
	userID, channelID := ctx.User().ID, ctx.ChannelID()
	if err := ctx.DeferEphemeral(); err != nil {
		return err
	}

	type expect struct {
		message *bot.MessageContext
		err     error
	}

	await, cancel := context.WithTimeout(ctx.Context(), 1*time.Minute)
	defer cancel()
	ch := make(chan expect, 1)
	go func() {
		message, err := ctx.Bot.AwaitMessage(await, func(mc *bot.MessageContext) bool {
			return mc.Author.ID == userID &&
				mc.ChannelID() == channelID &&
				len(mc.Attachments) > 0 || strings.TrimSpace(mc.Content) != ""
		})
		ch <- expect{message, err}
	}()

	if _, err := ctx.EditReply("Envie o JSON como arquivo ou cole o conteúdo JSON nesta conversa."); err != nil {
		return err
	}

	var received *bot.MessageContext
	select {
	case res := <-ch:
		if res.err != nil {
			return res.err
		}
		received = res.message
	case <-await.Done():
		return await.Err()
	}

	switch {
	case len(received.Attachments) > 0:
		if err := service.HandlerAttachments(received); err != nil {
			return err
		}
	case len(received.Content) > 0:
		if err := service.HandlerMessageContent(received); err != nil {
			return err
		}
	default:
		return nil
	}

	embed := messages.NewEmbedBuilder().
		SetTitle("𖥻 ׁ ׅ  Tudo certo e criado... ! ﹒  ◡◡").
		SetColor(0xfdfd96).
		SetTimestampNow().
		Build()

	if _, err := ctx.FollowupComplex(rest.ExecuteWebhookParams{
		Embeds: []messages.Embed{embed},
		Flags:  messages.FlagEphemeral,
	}); err != nil {
		return err
	}

	return nil
}

func (service *Service) EmbedsDelete(ctx *bot.InteractionContext) error {
	menu, err := service.GetMenuID(ctx)
	if err != nil {
		return err
	}

	if err := menu.interaction.DeferEphemeral(); err != nil {
		return err
	}

	data, err := service.database.DeleteOne(menu.interaction.Context(), menu.id)
	if err != nil {
		return err
	}

	embed := messages.NewEmbedBuilder().
		SetTitle("𖥻 ׁ ׅ  Incorporação deletada... ! ﹒  ◡◡").
		SetTimestampNow().
		SetColor(0xfdfd96)

	fields := map[string]string{
		"Criada por:":    fmt.Sprintf("<@%s>", data.UserID),
		"Identificador:": fmt.Sprintf("**%d**", data.ID),
	}

	for index, value := range fields {
		embed.AddField(index, value, true)
	}

	message := embed.Build()

	if _, err = ctx.FollowupComplex(rest.ExecuteWebhookParams{
		Embeds: []messages.Embed{message},
	}); err != nil {
		return err
	}
	return nil
}

func (service *Service) EmbedsObtain(ctx *bot.InteractionContext) error {
	menu, err := service.GetMenuID(ctx)
	if err != nil {
		return err
	}

	embed, err := service.database.GetOne(menu.interaction.Context(), menu.id)
	if err != nil {
		return err
	}

	var decode domain.EmbedContentDecode
	if err := json.Unmarshal([]byte(embed.Content), &decode); err != nil {
		return err
	}

	if len(decode.Content) > 0 {
		if _, err := menu.interaction.FollowupComplex(rest.ExecuteWebhookParams{
			Content: decode.Content,
			Embeds:  decode.Embeds,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (service *Service) EmbedsPublish(ctx *bot.InteractionContext) error {
	menu, err := service.GetMenuID(ctx)
	if err != nil {
		return err
	}
	if err := menu.interaction.DeferEphemeral(); err != nil {
		return err
	}

	channelMenuID := uuid.NewString()
	channelMenu := components.NewChannelSelectMenuBuilder().
		SetCustomID(channelMenuID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione qual canal deseja enviar").
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText))
	row := components.NewActionRowBuilder().AddComponents(channelMenu)
	if _, err := menu.interaction.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		return err
	}

	await, cancel := context.WithTimeout(menu.interaction.Context(), 1*time.Minute)
	defer cancel()
	interaction, err := menu.interaction.Bot.AwaitInteraction(await, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == channelMenuID &&
			ic.User().ID == ctx.User().ID
	})
	if err != nil {
		return err
	}

	if err := interaction.DeferEphemeral(); err != nil {
		return err
	}

	channelID, err := snowflake.Parse(interaction.Values()[0])
	if err != nil {
		return err
	}

	data, err := service.database.GetOne(interaction.Context(), menu.id)
	if err != nil {
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

	embed := messages.NewEmbedBuilder().SetTitle("𖥻 ׁ ׅ  Enviado com sucesso... ! ﹒  ◡◡").SetTimestampNow().SetColor(0xfdfd96).Build()
	if _, err := interaction.FollowupComplex(rest.ExecuteWebhookParams{
		Embeds: []messages.Embed{embed},
		Flags:  messages.FlagEphemeral,
	}); err != nil {
		return err
	}

	return nil
}

func (service *Service) HandlerAttachments(received *bot.MessageContext) error {
	return nil
}

func (service *Service) HandlerMessageContent(received *bot.MessageContext) error {
	formatted := strings.TrimSpace(received.Content)
	if !json.Valid([]byte(formatted)) {
		return errors.New("invalid json")
	}

	var embedContent domain.EmbedContentDecode

	if err := json.Unmarshal([]byte(formatted), &embedContent); err != nil {
		return err
	}
	bytes, err := json.Marshal(embedContent)
	if err != nil {
		return err
	}

	embed := domain.Embed{
		GuildID:   received.GuildID.String(),
		UserID:    received.Author.ID.String(),
		Content:   string(bytes),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if _, err := service.database.CreateOne(received.Context(), embed); err != nil {
		return err
	}

	return nil
}
