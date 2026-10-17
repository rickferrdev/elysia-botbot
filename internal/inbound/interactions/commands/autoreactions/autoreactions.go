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
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type FxArgs struct {
	fx.In

	Responder       responder.Interface
	InteractionFlow interactionflow.Interface
	AutoReactions   autoreactions.Interface
	DiscordClient   *discordclient.DiscordClient
}

type Command struct {
	responder     responder.Interface
	flow          interactionflow.Interface
	autoreactions autoreactions.Interface
	client        *discordclient.DiscordClient
}

func New(args FxArgs) *Command {
	server := Command{
		responder:     args.Responder,
		flow:          args.InteractionFlow,
		client:        args.DiscordClient,
		autoreactions: args.AutoReactions,
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

func (cmd *Command) execute(ctx *bot.InteractionContext) {
	cmd.AutoReactionsSetup(ctx)
}

func (cmd *Command) AutoReactionsSetup(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	menuID := uuid.NewString()
	interaction, err := cmd.flow.AwaitStringSelect(ctx, []components.SelectOption{
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

	channelSelectID := uuid.NewString()
	interaction, err := cmd.flow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		SetPlaceholder("Selecione um canal para monitorar.").
		Build())
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	channelID := interaction.Values()[0]

	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err = cmd.flow.AwaitSimpleTextModal(interaction, components.NewTextInputBuilder().
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
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	emoji := interaction.ModalValue(inputID)

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	stringSelectMenuID := uuid.NewString()
	interaction, err = cmd.flow.AwaitStringSelect(interaction, []components.SelectOption{
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
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	attachmentType := interaction.Values()[0]

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	data := domain.AutoReactions{
		GuildID:   interaction.Guild.ID.String(),
		ChannelID: channelID,
		Type:      domain.AutoReactionType(attachmentType),
		Emoji:     emoji,
	}

	result, err := cmd.autoreactions.CreateOne(interaction.Context(), data)
	if err != nil {
		if errors.Is(err, autoreactions.ErrConflict) {
			cmd.responder.Sender(ctx, responder.WithConflict, err)
			return
		}
		cmd.responder.Sender(ctx, responder.WithInternal, err)
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

	cmd.responder.Sender(interaction, responder.WithEmbed(embeder.Build()), nil)
}

func (cmd *Command) Obtain(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	menuID := uuid.NewString()
	interaction, err := cmd.flow.AwaitChannelSelect(ctx, components.NewChannelSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione o canal monitorado.").
		SetMaxValues(1).
		SetMinValues(1).
		SetChannelTypes(components.ChannelType(channels.ChannelTypeGuildText)).
		Build())
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	channelID := interaction.Values()[0]

	result, err := cmd.autoreactions.ObtainOne(ctx.Context(), interaction.Guild.ID.String(), channelID)
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			cmd.responder.Sender(ctx, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(ctx, responder.WithInternal, err)
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

	cmd.responder.Sender(interaction, responder.WithEmbed(embeder.Build()), nil)
}

func (cmd *Command) ObtainAll(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	data, err := cmd.autoreactions.ObtainAll(ctx.Context(), ctx.GuildID().String())
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			cmd.responder.Sender(ctx, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(ctx, responder.WithInternal, err)
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

	cmd.responder.Sender(ctx, responder.WithEmbed(embeder.Build()), nil)
}

func (cmd *Command) Update(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	channelSelectID := uuid.NewString()
	interaction, err := cmd.flow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione o canal monitorado.").
		Build())
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	channelID := interaction.Values()[0]

	menuID, inputID := uuid.NewString(), uuid.NewString()
	interaction, err = cmd.flow.AwaitSimpleTextModal(interaction, components.NewTextInputBuilder().
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
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	emoji := interaction.ModalValue(inputID)

	result, err := cmd.autoreactions.UpdateOne(ctx.Context(), interaction.Guild.ID.String(), channelID, emoji)
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			cmd.responder.Sender(ctx, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(ctx, responder.WithInternal, err)
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

	cmd.responder.Sender(interaction, responder.WithEmbed(embeder.Build()), nil)
}

func (cmd *Command) Delete(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	channelSelectID := uuid.NewString()
	interaction, err := cmd.flow.AwaitChannelSelect(ctx, components.NewChannelSelectBuilder().
		SetCustomID(channelSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione o canal monitorado.").
		Build())
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	channelID := interaction.Values()[0]

	result, err := cmd.autoreactions.DeleteOne(ctx.Context(), interaction.Guild.ID.String(), channelID)
	if err != nil {
		if errors.Is(err, autoreactions.ErrNotFound) {
			cmd.responder.Sender(ctx, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(ctx, responder.WithInternal, err)
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

	cmd.responder.Sender(interaction, responder.WithEmbed(embeder.Build()), nil)
}
