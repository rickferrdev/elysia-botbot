package server

import (
	"errors"
	"log/slog"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/server"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		flow      interactionflow.Interface
		responder responder.Interface
		storage   server.Interface
	}
	Params struct {
		fx.In
		InteractionFlow interactionflow.Interface
		Responder       responder.Interface
		Logger          *slog.Logger
		ServerStorage   server.Interface
	}
)

type Interface interface {
	ServerSetup(ctx *bot.InteractionContext, router *bot.Router)
}

func New(params Params) *Service {
	service := Service{
		flow:      params.InteractionFlow,
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
	interaction, err := service.flow.AwaitStringSelect(
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
			SetPlaceholder("Selecione uma opção").
			SetMinValues(1).
			SetMaxValues(1).
			Build(),
	)
	if err != nil {
		service.responder.SendError(ctx, err)
		return
	}

	switch interaction.Values()[0] {
	case "obtain":
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
	if err := ctx.DeferUpdate(); err != nil {
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
		Title:            "Ok",
		WithTimestampNow: true,
		WithColor:        true,
	})

	return nil
}

func (service *Service) Create(ctx *bot.InteractionContext) error {
	inputID := uuid.NewString()
	modalID := uuid.NewString()
	interaction, err := service.flow.AwaitSimpleTextModal(ctx,
		components.NewTextInputBuilder().
			SetCustomID(inputID).
			SetLabel("Digite um prefixo para o bot").
			SetStyle(components.TextInputStyleShort).
			SetPlaceholder("e!").
			SetRequired(true).
			SetMaxLength(1).
			Build(),
		components.NewModalBuilder().SetCustomID(modalID).SetTitle("Adicionar/Atualizar as configurações").Build(),
	)
	if err != nil {
		return err
	}

	if err := interaction.DeferUpdate(); err != nil {
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
		Title:            "Ok",
		WithTimestampNow: true,
		WithColor:        true,
	})

	return nil
}
