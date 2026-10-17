package server

import (
	"errors"
	"fmt"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/users"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/server"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type FxArgs struct {
	fx.In

	DiscordClient   *discordclient.DiscordClient
	Responder       responder.Interface
	Server          server.Interface
	InteractionFlow interactionflow.Interface
}

type Command struct {
	flow      interactionflow.Interface
	client    *discordclient.DiscordClient
	server    server.Interface
	responder responder.Interface
}

func New(args FxArgs) *Command {
	server := Command{
		flow:      args.InteractionFlow,
		responder: args.Responder,
		client:    args.DiscordClient,
		server:    args.Server,
	}

	args.DiscordClient.Router.Command(
		"server",
		"Configurações úteis para o servidor.",
		server.execute,
	).
		Use(bot.RequirePermissions(
			permissions.Build(
				permissions.Administrator,
			),
		))
	return &server
}

func (cmd *Command) execute(ctx *bot.InteractionContext) {
	cmd.ServerSetup(ctx, cmd.client.Router)

	cmd.responder.Sender(ctx, responder.WithOk, nil)

}

func (cmd *Command) ServerSetup(ctx *bot.InteractionContext, router *bot.Router) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	menuID := uuid.NewString()
	interaction, err := cmd.flow.AwaitStringSelect(
		ctx,
		[]components.SelectOption{
			components.NewSelectOptionBuilder().
				SetLabel("Obter configurações").
				SetValue("obtain").
				SetDescription("Consultar as configurações atuais do servidor").
				Build(),
			components.NewSelectOptionBuilder().
				SetLabel("Criar/Atualizar configurações").
				SetValue("upsert").
				SetDescription("Criar ou atualizar as configurações do servidor").
				Build(),
		}, components.NewStringSelectMenuBuilder().
			SetCustomID(menuID).
			SetPlaceholder("Selecione uma opção.").
			SetMinValues(1).
			SetMaxValues(1).
			Build(),
	)
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	switch interaction.Values()[0] {
	case "obtain":
		cmd.Obtain(interaction)
		return
	case "upsert":
		cmd.Create(interaction)
		return
	default:
		return
	}
}

func (cmd *Command) Obtain(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	data, err := cmd.server.ObtainServerWithConfig(ctx.Context(), ctx.Guild.ID.String())
	if err != nil {
		if errors.Is(err, server.ErrNotFound) {
			cmd.responder.Sender(ctx, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}
	if data.Config == nil {
		// TODO: atualizar para nova assinatura do responder (WithSimpleCustomMessage não existe mais)
		// cmd.responder.Sender(responder.WithSimpleCustomMessage(ctx, "Solicite a configuração do servidor antes.", "", nil))
		return
	}

	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Configurações do Servidor 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	fields := map[string]string{
		"ִ ࣪𖤐 Identificador do Servidor:": fmt.Sprintf("**%v**", data.GuildID),
		"ִ ࣪𖤐 Prefixo do Elysia-BotBot:":  fmt.Sprintf("`%v`", data.Config.Prefix),
	}

	for index, value := range fields {
		embeder.AddField(index, value, false)
	}

	cmd.responder.Sender(ctx, responder.WithEmbed(embeder.Build()), nil)
}

func (cmd *Command) Create(ctx *bot.InteractionContext) {
	inputID := uuid.NewString()
	modalID := uuid.NewString()
	interaction, err := cmd.flow.AwaitSimpleTextModal(ctx,
		components.NewTextInputBuilder().
			SetCustomID(inputID).
			SetLabel("Digite um prefixo para o bot").
			SetStyle(components.TextInputStyleShort).
			SetPlaceholder("e!").
			SetRequired(true).
			SetMaxLength(5).
			Build(),
		components.NewModalBuilder().SetCustomID(modalID).SetTitle("Adicionar/Atualizar as configurações").Build(),
	)
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	err = cmd.server.CreateServerWithConfig(interaction.Context(), domain.Server{
		GuildID: interaction.Guild.ID.String(),
		Config: &domain.Config{
			GuildID: interaction.Guild.ID.String(),
			Prefix:  interaction.ModalValue(inputID),
		},
	})
	if err != nil {
		if errors.Is(err, server.ErrConflict) {
			cmd.responder.Sender(ctx, responder.WithConflict, err)
			return
		}
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	embeder := messages.NewEmbedBuilder().
		SetTitle("⋆. 𐙚˚࿔ Configurações do Servidor 𝜗𝜚˚⋆").
		SetThumbnail(ctx.Bot.User().
			DisplayAvatarURL(users.AvatarURLOptions{})).
		SetColor(0xB4E4D6).
		SetTimestampNow()

	fields := map[string]string{
		"ִ ࣪𖤐 Prefixo do Elysia-BotBot:": fmt.Sprintf("`%v`", interaction.ModalValue(inputID)),
	}

	for index, value := range fields {
		embeder.AddField(index, value, false)
	}

	cmd.responder.Sender(interaction, responder.WithEmbed(embeder.Build()), nil)
}
