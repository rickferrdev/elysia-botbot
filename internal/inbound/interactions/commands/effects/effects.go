package effects

import (
	"errors"
	"fmt"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/effects"
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

		KwaiiService  effects.Interface
		DiscordClient *discordclient.DiscordClient
		Responder     responder.Interface
		Flow          interactionflow.Interface
	}

	Command struct {
		service   effects.Interface
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
		"effects",
		"Gerencie os efeitos ativos no servidor.",
		cmd.execute,
	).Use(bot.RequirePermissions(permissions.Build(permissions.Administrator)))

	return &cmd
}

func (cmd *Command) execute(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	menuID := uuid.NewString()
	interaction, err := cmd.flow.AwaitStringSelect(ctx, []components.SelectOption{
		components.NewSelectOptionBuilder().
			SetLabel("Criar").
			SetValue("create").
			SetDescription("Aplicar um efeito a um membro").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Listar").
			SetValue("list").
			SetDescription("Listar os efeitos ativos no servidor").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Atualizar").
			SetValue("update").
			SetDescription("Alterar o efeito de um membro").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Deletar").
			SetValue("delete").
			SetDescription("Remover o efeito de um membro").
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

	if len(interaction.Values()) != 1 {
		cmd.responder.Sender(interaction, responder.WithBadArguments, errors.New("invalid effect action selection"))
		return
	}

	switch interaction.Values()[0] {
	case "create":
		cmd.CreateOne(interaction)
	case "list":
		cmd.GetAll(interaction)
	case "update":
		cmd.UpdateOne(interaction)
	case "delete":
		cmd.DeleteOne(interaction)
	default:
		cmd.responder.Sender(interaction, responder.WithBadArguments, errors.New("unknown effect action"))
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
		SetPlaceholder("Selecione o membro que receberá o efeito.").
		Build())
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	userID, err := cmd.selectedUserID(interaction)
	if err != nil {
		cmd.responder.Sender(interaction, responder.WithBadArguments, err)
		return
	}

	effectInteraction, effectType, err := cmd.selectEffect(interaction)
	if err != nil {
		if effectInteraction == nil {
			cmd.responder.Sender(interaction, responder.WithInternal, err)
			return
		}
		cmd.responder.Sender(effectInteraction, responder.WithBadArguments, err)
		return
	}

	if err := effectInteraction.DeferUpdate(); err != nil {
		cmd.responder.Sender(effectInteraction, responder.WithInternal, err)
		return
	}

	now := time.Now().UTC()
	result, err := cmd.service.CreateOne(effectInteraction.Context(), domain.Effect{
		UserID:     userID,
		GuildID:    effectInteraction.Guild.ID.String(),
		Type:       effectType,
		Expiration: now.Add(30 * time.Minute),
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		if errors.Is(err, effects.ErrEffectAlreadyExists) {
			cmd.responder.Sender(effectInteraction, responder.WithConflict, err)
			return
		}
		cmd.responder.Sender(effectInteraction, responder.WithInternal, err)
		return
	}

	avatar := effectInteraction.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})
	cmd.responder.Sender(effectInteraction, responder.WithEmbed(
		resembeds.EmbedCreateOneKwaii(avatar, *result),
	), nil)
}

func (cmd *Command) GetAll(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	result, err := cmd.service.ObtainAll(ctx.Context(), ctx.GuildID().String())
	if err != nil {
		if errors.Is(err, effects.ErrEffectNotFound) {
			cmd.responder.Sender(ctx, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}
	if result == nil {
		cmd.responder.Sender(ctx, responder.WithInternal, errors.New("effect service returned no list"))
		return
	}

	avatar := ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})
	cmd.responder.Sender(ctx, responder.WithEmbed(
		resembeds.EmbedObtainAllKwaii(avatar, *result),
	), nil)
}

func (cmd *Command) UpdateOne(ctx *bot.InteractionContext) {
	if err := ctx.DeferUpdate(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	userSelectID := uuid.NewString()
	interaction, err := cmd.flow.AwaitUserSelect(ctx, components.NewUserSelectBuilder().
		SetCustomID(userSelectID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione o membro cujo efeito será atualizado.").
		Build())
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	userID, err := cmd.selectedUserID(interaction)
	if err != nil {
		cmd.responder.Sender(interaction, responder.WithBadArguments, err)
		return
	}

	effectInteraction, effectType, err := cmd.selectEffect(interaction)
	if err != nil {
		if effectInteraction == nil {
			cmd.responder.Sender(interaction, responder.WithInternal, err)
			return
		}
		cmd.responder.Sender(effectInteraction, responder.WithBadArguments, err)
		return
	}

	if err := effectInteraction.DeferUpdate(); err != nil {
		cmd.responder.Sender(effectInteraction, responder.WithInternal, err)
		return
	}

	now := time.Now().UTC()
	result, err := cmd.service.UpdateOne(effectInteraction.Context(), domain.Effect{
		UserID:     userID,
		GuildID:    effectInteraction.Guild.ID.String(),
		Type:       effectType,
		Expiration: now.Add(30 * time.Minute),
		UpdatedAt:  now,
	})
	if err != nil {
		if errors.Is(err, effects.ErrEffectNotFound) {
			cmd.responder.Sender(effectInteraction, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(effectInteraction, responder.WithInternal, err)
		return
	}

	avatar := effectInteraction.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})
	cmd.responder.Sender(effectInteraction, responder.WithEmbed(
		resembeds.EmbedUpdateOneKwaii(avatar, *result),
	), nil)
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
		SetPlaceholder("Selecione o membro cujo efeito será removido.").
		Build())
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	userID, err := cmd.selectedUserID(interaction)
	if err != nil {
		cmd.responder.Sender(interaction, responder.WithBadArguments, err)
		return
	}

	result, err := cmd.service.DeleteOne(interaction.Context(), interaction.GuildID().String(), userID)
	if err != nil {
		if errors.Is(err, effects.ErrEffectNotFound) {
			cmd.responder.Sender(interaction, responder.WithNotFound, err)
			return
		}
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	avatar := interaction.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})
	cmd.responder.Sender(interaction, responder.WithEmbed(
		resembeds.EmbedDeleteOneKwaii(avatar, *result),
	), nil)
}

func (cmd *Command) selectEffect(ctx *bot.InteractionContext) (*bot.InteractionContext, domain.EffectType, error) {
	menuID := uuid.NewString()
	interaction, err := cmd.flow.AwaitStringSelect(ctx, []components.SelectOption{
		components.NewSelectOptionBuilder().
			SetLabel("Kwaii").
			SetValue(string(domain.EffectTypeKwaii)).
			SetDescription("Transforma as mensagens do membro em falas kawaii.").
			Build(),
		components.NewSelectOptionBuilder().
			SetLabel("Christmas").
			SetDescription("Transforma as mensagens dos membros em falas natalinas.").
			SetValue("christmas").
			Build(),
	}, components.NewStringSelectMenuBuilder().
		SetCustomID(menuID).
		SetMaxValues(1).
		SetMinValues(1).
		SetPlaceholder("Selecione o efeito.").
		Build())
	if err != nil {
		return nil, "", err
	}
	if len(interaction.Values()) != 1 || interaction.Values()[0] == "" {
		return interaction, "", errors.New("invalid effect type selection")
	}

	return interaction, domain.EffectType(interaction.Values()[0]), nil
}

func (cmd *Command) selectedUserID(ctx *bot.InteractionContext) (string, error) {
	values := ctx.Values()
	if len(values) != 1 {
		return "", errors.New("invalid user selection")
	}

	userID, err := snowflake.Parse(values[0])
	if err != nil {
		return "", fmt.Errorf("invalid selected user ID: %w", err)
	}

	return userID.String(), nil
}
