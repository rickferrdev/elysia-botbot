package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/rest"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/server"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		responder responder.Interface
		storage   server.Interface
	}
	Params struct {
		fx.In
		Responder     responder.Interface
		Logger        *slog.Logger
		ServerStorage server.Interface
	}
)

type Interface interface {
	ServerSetup(ctx *bot.InteractionContext, router *bot.Router)
}

func New(params Params) *Service {
	service := Service{
		responder: params.Responder,
		storage:   params.ServerStorage,
	}

	return &service
}

func (service *Service) ServerSetup(ctx *bot.InteractionContext, router *bot.Router) {
	if err := ctx.DeferEphemeral(); err != nil {
		service.responder.SendError(ctx, err)
		return
	}

	menuID := uuid.NewString()
	options := []components.SelectOption{
		components.NewSelectOptionBuilder().
			SetLabel("Obter configurações").
			SetValue("read").
			SetDescription("Consultar as configurações atuais do servidor").
			Build(),

		components.NewSelectOptionBuilder().
			SetLabel("Criar/Atualizar configurações").
			SetValue("upsert").
			SetDescription("Criar ou atualizar as configurações do servidor").
			Build(),
	}

	stringSelectMenu := components.NewStringSelectMenuBuilder().
		SetCustomID(menuID).
		SetPlaceholder("Selecione uma opção").
		AddOptions(options[0], options[1]).
		SetMinValues(1).
		SetMaxValues(1).
		Build()

	row := components.NewActionRowBuilder().AddComponents(stringSelectMenu).Build()
	if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		service.responder.SendError(ctx, err)
		return
	}

	await, cancel := context.WithTimeout(ctx.Context(), 2*time.Minute)
	defer cancel()
	interaction, err := ctx.Bot.AwaitInteraction(await, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == menuID && ic.User().ID == ctx.User().ID
	})
	if err != nil {
		service.responder.SendError(ctx, err)
		return
	}

	switch interaction.Values()[0] {
	case "read":
		if err := service.Obtain(interaction); err != nil {
			service.responder.SendError(interaction, err)
			return
		}
		return
	case "upsert":
		if err := service.Obtain(interaction); err != nil {
			service.responder.SendError(interaction, err)
			return
		}
		return
	default:
		return
	}
}

func (service *Service) Obtain(ctx *bot.InteractionContext) error {
	if err := ctx.DeferEphemeral(); err != nil {
		return err
	}

	data, err := service.storage.GetServerWithConfig(ctx.Context(), ctx.Guild.ID.String())
	if err != nil {
		return err
	}
	if data.Config == nil {
		return errors.New("config not found")
	}

	service.responder.SendMessage(ctx, responder.Message{
		Title:            fmt.Sprintf("**Identificador:** %d", data.ID),
		WithTimestampNow: true,
		WithColor:        true,
	})

	return nil
}

func (service *Service) Create(ctx *bot.InteractionContext) error {
	inputID := uuid.NewString()
	modalID := uuid.NewString()

	input := components.NewTextInputBuilder().
		SetCustomID(inputID).
		SetLabel("Digite um prefixo para o bot").
		SetStyle(components.TextInputStyleShort).
		SetPlaceholder("e!").
		SetRequired(true).
		SetMaxLength(1).
		Build()

	row := components.NewActionRowBuilder().AddComponents(input)
	if err := ctx.ShowModal(modalID, "Adicionar/Atualizar as configurações", row); err != nil {
		return err
	}

	await, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	interaction, err := ctx.Bot.AwaitInteraction(await, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == modalID && ic.User().ID == ctx.User().ID
	})
	if err != nil {
		return err
	}

	if err := interaction.DeferEphemeral(); err != nil {
		return err
	}

	err = service.storage.CreateServerWithConfig(interaction.Context(), domain.Server{
		GuildID: interaction.Guild.ID.String(),
		Config: &domain.Config{
			GuildID: interaction.Guild.ID.String(),
			Prefix:  interaction.ModalValue(inputID),
		},
	})
	if err != nil {
		return err
	}

	service.responder.SendMessage(interaction, responder.Message{
		Title:            "Operação concluída",
		WithTimestampNow: true,
		WithColor:        true,
	})

	return nil
}
