package server

import (
	"log/slog"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/server"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		flow      interactionflow.Interface
		responder responder.Interface
		storage   server.Interface
		logger    *slog.Logger
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
		logger:    params.Logger,
	}

	return &service
}

func (service *Service) ServerSetup(ctx *bot.InteractionContext, router *bot.Router) {
	if err := ctx.DeferEphemeral(); err != nil {
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
		service.responder.
			Sender(responder.WithStatusInternalServerError(ctx, err))
		return
	}

	switch interaction.Values()[0] {
	case "obtain":
		if err := service.Obtain(interaction); err != nil {
			service.responder.Sender(
				responder.WithStatusInternalServerError(ctx, err),
			)
		}
		return
	case "upsert":
		if err := service.Create(interaction); err != nil {
			service.responder.Sender(
				responder.WithStatusInternalServerError(ctx, err),
			)
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
		if core.IsCode(err, repoerrors.CodeSqlNoRows) {
			service.responder.Sender(responder.WithStatusNotFound(ctx, err))
			return nil
		}
		return err
	}
	if data.Config == nil {
		service.responder.Sender(responder.WithCustomMessage(ctx, "config does not exist", "", nil))
		return nil
	}

	service.responder.Sender(responder.WithStatusOk(ctx, nil))
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
		if core.IsCode(err, core.CodeConflict) {
			service.responder.Sender(responder.WithStatusConflict(interaction, err))
			return nil
		}
		return err
	}

	service.responder.Sender(responder.WithStatusCreated(interaction, nil))
	return nil
}
