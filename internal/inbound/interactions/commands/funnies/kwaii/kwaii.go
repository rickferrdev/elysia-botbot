package kwaii

import (
	"errors"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/funnies/kwaii"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/resembeds"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type (
	FxArgs struct {
		fx.In

		KwaiiService  kwaii.Interface
		DiscordClient *discordclient.DiscordClient
		Responder     responder.Interface
		Flow          interactionflow.Interface
	}

	Command struct {
		service   kwaii.Interface
		flow      interactionflow.Interface
		client    *discordclient.DiscordClient
		responder responder.Interface
	}
)

func New(args FxArgs) *Command {
	cmd := Command{
		flow:      args.Flow,
		service:   args.KwaiiService,
		client:    args.DiscordClient,
		responder: args.Responder,
	}

	cmd.client.Router.Command(
		"kwaii",
		"Manipule os efeitos kwaii!",
		cmd.execute,
	).RequireBotPermissions(permissions.AddReactions)
	return &cmd
}

func (cmd *Command) execute(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	menuID := uuid.NewString()
	interaction, err := cmd.flow.AwaitStringSelect(ctx, []components.SelectOption{
		components.NewSelectOptionBuilder().
			SetLabel("Adicionar").
			SetValue("create").
			SetDescription("Adicionar um novo efeito kwaii").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Deletar").
			SetValue("delete").
			SetDescription("Deletar um efeito kwaii").
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
		cmd.CreateOne(interaction)
		return
	case "delete":
		cmd.DeleteOne(interaction)
		return
	default:
		return
	}
}

func (cmd *Command) CreateOne(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	userSelectID := uuid.NewString()
	interaction, err := cmd.flow.AwaitUserSelect(ctx, components.NewUserSelectBuilder().
		SetCustomID(userSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione qual usuário aplicar o efeito.").
		Build())
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if interaction.Values()[0] == "" {
		cmd.responder.Sender(ctx, responder.WithBadArguments, err)
		return
	}

	userID, err := snowflake.Parse(interaction.Values()[0])
	if err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	if err := cmd.service.CreateOne(interaction.Context(), userID.String()); err != nil {
		if errors.Is(err, kwaii.ErrKwaiiEffectAlreadyExists) {
			cmd.responder.Sender(interaction, responder.WithConflict, err)
			return
		}
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	embed := resembeds.EmbedCreateOneKwaii()
	cmd.responder.Sender(interaction, responder.WithEmbed(embed), nil)
}

func (cmd *Command) DeleteOne(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	userSelectID := uuid.NewString()
	interaction, err := cmd.flow.AwaitUserSelect(ctx, components.NewUserSelectBuilder().
		SetCustomID(userSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione qual usuário aplicar o efeito.").
		Build())
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if interaction.Values()[0] == "" {
		cmd.responder.Sender(ctx, responder.WithBadArguments, err)
		return
	}

	userID, err := snowflake.Parse(interaction.Values()[0])
	if err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	if err := cmd.service.DeleteOne(interaction.Context(), userID.String()); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	embed := resembeds.EmbedDeleteOneKwaii()
	cmd.responder.Sender(interaction, responder.WithEmbed(embed), nil)
}
