package autoreactions

import (
	"errors"
	"fmt"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/users"
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
	responder            responder.Interface
	interactionFlow      interactionflow.Interface
	autoReactionsService autoreactions.Interface
	discordClient        *discord.DiscordClient
}

func New(args Args) *Command {
	server := Command{
		responder:            args.ResponderService,
		interactionFlow:      args.InteractionFlow,
		discordClient:        args.DiscordClient,
		autoReactionsService: args.AutoReactionsService,
	}
	args.DiscordClient.Router.Command(
		"autoreactions",
		"Configurações úteis relacionadas a autorreações.",
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
			SetDescription("Criar um novo registro de autorreação").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Obter").
			SetValue("obtain").
			SetDescription("Consultar autorreações e seus anexos").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Obter tudo").
			SetValue("obtain-all").
			SetDescription("Obtém todas as autorreações existentes").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Atualizar").
			SetValue("update").
			SetDescription("Atualizar uma autorreação existente").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Deletar").
			SetValue("delete").
			SetDescription("Remover uma autorreação existente").
			Build(),
	}, components.NewStringSelectMenuBuilder().
		SetCustomID(menuID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione uma opção.").
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
	case "update":
		command.Update(interaction)
		return
	case "delete":
		command.Delete(interaction)
		return
	case "obtain-all":
		command.ObtainAll(interaction)
	default:
		return
	}
}
func (command *Command) Create(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	channelSelectID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		SetPlaceholder("Selecione um canal para monitorar.").
		Build())
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	channelID := interaction.Values()[0]
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err = command.interactionFlow.AwaitSimpleTextModal(interaction, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetPlaceholder("Qual emoji usar? Ex.: 🐶").
		SetStyle(components.TextInputStyleShort).
		SetRequired(true).
		SetLabel("Emoji").
		Build(), components.NewModalBuilder().
		SetTitle("Configurações").
		SetCustomID(menuID).
		Build())
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	emoji := interaction.ModalValue(inputID)
	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	stringSelectMenuID := uuid.NewString()
	interaction, err = command.interactionFlow.AwaitStringSelect(interaction, []components.SelectOption{
		components.NewSelectOptionBuilder().
			SetLabel("Mensagem").
			SetValue(string(domain.AutoReactionTypeMessage)).
			SetDescription("Monitorar as mensagens recebidas no canal.").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Anexos").
			SetValue(string(domain.AutoReactionTypeAttachment)).
			SetDescription("Monitorar os anexos recebidos no canal.").
			Build(),
	}, components.NewStringSelectMenuBuilder().
		SetCustomID(stringSelectMenuID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione o que deve ser monitorado.").
		Build())
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	attachmentType := interaction.Values()[0]
	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	data := domain.AutoReactions{
		GuildID:   interaction.Guild.ID.String(),
		ChannelID: channelID,
		Type:      domain.AutoReactionType(attachmentType),
		Emoji:     emoji,
	}
	result, err := command.autoReactionsService.CreateOne(interaction.Context(), data)
	if err != nil {
		if errors.Is(err, autoreactions.ErrConflict) {
			command.responder.Sender(responder.WithStatusConflict(ctx, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Criação de AutoReaction 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()
	fields := map[string]string{
		"── .✦ Tipo de monitoramento:": fmt.Sprintf("`%v`", result.Type),
		"⋆˚꩜｡ Canal registrado:":       fmt.Sprintf("<#%v>", result.ChannelID),
		"⋆˚꩜｡ Emoji:":                  fmt.Sprintf("%v", result.Emoji),
	}
	for index, value := range fields {
		embeder.AddField(index, value, false)
	}
	command.responder.Sender(responder.WithCustomEmbed(interaction, embeder.Build(), nil))
}
func (command *Command) Obtain(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	menuID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitChannelSelect(ctx, components.NewChannelSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione o canal monitorado.").
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		Build())
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	channelID := interaction.Values()[0]
	result, err := command.autoReactionsService.ObtainOne(ctx.Context(), interaction.Guild.ID.String(), channelID)
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Obter uma AutoReaction 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()
	fields := map[string]string{
		"── .✦ Tipo de monitoramento:": fmt.Sprintf("`%v`", result.Type),
		"⋆˚꩜｡ Canal registrado:":       fmt.Sprintf("<#%v>", result.ChannelID),
		"⋆˚꩜｡ Emoji:":                  fmt.Sprintf("%v", result.Emoji),
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
	data, err := command.autoReactionsService.ObtainAll(ctx.Context(), ctx.GuildID().String())
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Obter todas AutoReactions 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()
	var fields map[string]string
	for _, value := range data {
		fields = map[string]string{
			fmt.Sprintf("<#%v> -> %v", value.ChannelID, value.Emoji): fmt.Sprintf("Tipo de monitoramento: `%v`", value.Type),
		}
	}
	for index, value := range fields {
		embeder.AddField(index, value, false)
	}
	command.responder.Sender(responder.WithCustomEmbed(ctx, embeder.Build(), nil))
}
func (command *Command) Update(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	channelSelectID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione o canal monitorado.").
		Build())
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	channelID := interaction.Values()[0]
	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err = command.interactionFlow.AwaitSimpleTextModal(interaction, components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetPlaceholder("Qual emoji usar? Ex.: 🐶").
		SetStyle(components.TextInputStyleShort).
		SetRequired(true).
		SetLabel("Emoji").
		Build(), components.NewModalBuilder().
		SetTitle("Configurações").
		SetCustomID(menuID).
		Build())
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	emoji := interaction.ModalValue(inputID)
	result, err := command.autoReactionsService.UpdateOne(ctx.Context(), interaction.Guild.ID.String(), channelID, emoji)
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Atualizar uma AutoReaction 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()
	fields := map[string]string{
		"── .✦ Tipo de monitoramento:": fmt.Sprintf("`%v`", result.Type),
		"⋆˚꩜｡ Canal registrado:":       fmt.Sprintf("<#%v>", result.ChannelID),
		"⋆˚꩜｡ Emoji:":                  fmt.Sprintf("%v", result.Emoji),
	}
	for index, value := range fields {
		embeder.AddField(index, value, false)
	}
	command.responder.Sender(responder.WithCustomEmbed(interaction, embeder.Build(), nil))
}
func (command *Command) Delete(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	channelSelectID := uuid.NewString()
	interaction, err := command.interactionFlow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione o canal monitorado.").
		Build())
	if err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	if err := interaction.DeferUpdate(); err != nil {
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	channelID := interaction.Values()[0]
	result, err := command.autoReactionsService.DeleteOne(ctx.Context(), interaction.Guild.ID.String(), channelID)
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			command.responder.Sender(responder.WithStatusNotFound(ctx, err))
			return
		}
		command.responder.Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}
	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Deletar uma AutoReaction 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	fields := map[string]string{
		"── .✦ Tipo de monitoramento:": fmt.Sprintf("`%v`", result.Type),
		"⋆˚꩜｡ Canal registrado:":       fmt.Sprintf("<#%v>", result.ChannelID),
		"⋆˚꩜｡ Emoji:":                  fmt.Sprintf("%v", result.Emoji),
	}
	for index, value := range fields {
		embeder.AddField(index, value, false)
	}
	command.responder.Sender(responder.WithCustomEmbed(interaction, embeder.Build(), nil))
}
