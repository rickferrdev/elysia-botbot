package responder

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
	"go.uber.org/fx"
)

var Provide = fx.Provide(
	fx.Annotate(New, fx.As(new(Interface))),
)

const (
	ColorSuccess = 0xA6E3A1 // Soft green
	ColorError   = 0xF38BA8 // Soft red
	ColorInfo    = 0x89B4FA // Soft blue
	ColorNeutral = 0xCDD6F4 // Soft white
)

type Interface interface {
	Sender(sender SenderFN) error
}

type Responder struct {
	logger *slog.Logger
}

type Params struct {
	fx.In

	Logger *slog.Logger
}

func New(params Params) *Responder {
	return &Responder{
		logger: params.Logger,
	}
}

type SenderFN func() (*ResponderData, error)

type ResponderData struct {
	interaction     *bot.InteractionContext
	embed           messages.Embed
	content         string
	err             error
	clearComponents bool
}

func (responder *Responder) Sender(sender SenderFN) error {
	if sender == nil {
		return errors.New("responder: sender cannot be nil")
	}

	data, err := sender()
	if err != nil {
		return fmt.Errorf("responder: failed to build response: %w", err)
	}

	if data == nil {
		return errors.New("responder: sender returned nil data")
	}

	sendErr := responder.HandleData(*data)

	if data.err != nil {
		if sendErr != nil {
			return fmt.Errorf(
				"responder: failed to send error response: %w",
				sendErr,
			)
		}

		return data.err
	}

	if sendErr != nil {
		return fmt.Errorf("responder: failed to send response: %w", sendErr)
	}

	return nil
}

func newSimpleEmbed(title, description string, color int, timestamp bool) messages.Embed {
	embed := messages.NewEmbedBuilder().
		SetTitle(title).
		SetDescription(description).
		SetColor(color)

	if timestamp {
		embed.SetTimestampNow()
	}

	return embed.Build()
}

func WithStatusOk(
	ctx *bot.InteractionContext,
	err error,
) SenderFN {
	return withStatus(
		ctx,
		"Concluído com sucesso!",
		"Sua solicitação foi processada. Tudo certo!",
		ColorSuccess,
		err,
	)
}

func WithStatusCreated(
	ctx *bot.InteractionContext,
	err error,
) SenderFN {
	return withStatus(
		ctx,
		"Criado com sucesso!",
		"Prontinho! O recurso foi criado e já está disponível.",
		ColorSuccess,
		err,
	)
}

func WithStatusNoContent(
	ctx *bot.InteractionContext,
	err error,
) SenderFN {
	return withStatus(
		ctx,
		"Tudo certo!",
		"Sua solicitação foi concluída com sucesso.",
		ColorNeutral,
		err,
	)
}

func WithStatusBadRequest(
	ctx *bot.InteractionContext,
	err error,
) SenderFN {
	return withStatus(
		ctx,
		"Ops! Algo não saiu como esperado",
		"Confira as informações enviadas e tente novamente. Se precisar, revise os campos e tente mais uma vez.",
		ColorError,
		err,
	)
}

func WithStatusUnauthorized(
	ctx *bot.InteractionContext,
	err error,
) SenderFN {
	return withStatus(
		ctx,
		"Precisamos confirmar sua identidade",
		"Você precisa estar autenticado para continuar. Faça a autenticação e tente novamente.",
		ColorError,
		err,
	)
}

func WithStatusForbidden(
	ctx *bot.InteractionContext,
	err error,
) SenderFN {
	return withStatus(
		ctx,
		"Acesso não permitido",
		"Você não tem permissão para realizar esta ação. Se acredita que isso é um engano, entre em contato com a equipe responsável.",
		ColorError,
		err,
	)
}

func WithStatusNotFound(
	ctx *bot.InteractionContext,
	err error,
) SenderFN {
	return withStatus(
		ctx,
		"Não encontramos o que você procura",
		"O item solicitado não foi encontrado. Confira as informações e tente novamente.",
		ColorInfo,
		err,
	)
}

func WithStatusConflict(
	ctx *bot.InteractionContext,
	err error,
) SenderFN {
	return withStatus(
		ctx,
		"Esta ação não pode ser concluída agora",
		"Já existe um registro ou uma situação que impede esta operação. Confira os dados e tente novamente.",
		ColorError,
		err,
	)
}

func WithStatusInternalServerError(
	ctx *bot.InteractionContext,
	err error,
) SenderFN {
	return withStatus(
		ctx,
		"Ops! Tivemos um imprevisto",
		"Não conseguimos concluir sua solicitação neste momento. Tente novamente em instantes. Se o problema continuar, avise nossa equipe.",
		ColorError,
		err,
	)
}

func WithCustomMessage(ctx *bot.InteractionContext, title, message string, cause error, clearComponents bool) SenderFN {
	return withStatus(ctx, title, message, ColorNeutral, cause)
}

func WithCustomEmbed(ctx *bot.InteractionContext, embed messages.Embed, cause error) SenderFN {
	return withEmbed(ctx, embed, cause)
}

func withStatus(
	ctx *bot.InteractionContext,
	title string,
	description string,
	color int,
	cause error,
) SenderFN {
	return func() (*ResponderData, error) {
		if ctx == nil {
			return nil, errors.New(
				"responder: interaction context cannot be nil",
			)
		}

		if !ctx.IsRepliable() {
			return nil, errors.New(
				"responder: interaction is not repliable",
			)
		}

		var err error
		if cause != nil {
			err = fmt.Errorf(
				"responder: status response %q: %w",
				title,
				cause,
			)
		}

		return &ResponderData{
			interaction: ctx,
			embed: newSimpleEmbed(
				title,
				description,
				color,
				true,
			),
			err: err,
		}, nil
	}
}

func withEmbed(ctx *bot.InteractionContext, embed messages.Embed, cause error) SenderFN {
	return func() (*ResponderData, error) {
		if ctx == nil {
			return nil, errors.New(
				"responder: interaction context cannot be nil",
			)
		}

		if !ctx.IsRepliable() {
			return nil, errors.New(
				"responder: interaction is not repliable",
			)
		}

		var err error
		if cause != nil {
			err = fmt.Errorf(
				"responder: status response %q: %w",
				embed.Title,
				cause,
			)
		}

		return &ResponderData{
			interaction: ctx,
			embed:       embed,
			err:         err,
		}, nil
	}
}

func (responder *Responder) HandleData(data ResponderData) error {
	if data.err != nil {
		responder.logger.Error(data.err.Error())
	}
	if data.interaction == nil {
		return errors.New("responder: interaction context cannot be nil")
	}

	if !data.interaction.IsRepliable() {
		return errors.New("responder: interaction is not repliable")
	}

	ctx := data.interaction

	if ctx.Replied() {
		_, err := ctx.EditReplyComplex(rest.EditMessageParams{
			Embeds:     &[]messages.Embed{data.embed},
			Components: &[]components.Component{},
		})
		if err != nil {
			responder.logger.Error(
				"failed to send interaction follow-up",
				slog.String("error", err.Error()),
			)

			return fmt.Errorf(
				"responder: failed to send interaction follow-up: %w",
				err,
			)
		}

		return nil
	}

	if ctx.Deferred() {
		params := rest.EditMessageParams{
			Embeds:     &[]messages.Embed{data.embed},
			Components: &[]components.Component{},
		}

		if data.content != "" {
			params.Content = &data.content
		}

		_, err := ctx.EditReplyComplex(params)
		if err != nil {
			responder.logger.Error(
				"failed to edit interaction response",
				slog.String("error", err.Error()),
			)

			return fmt.Errorf(
				"responder: failed to edit interaction response: %w",
				err,
			)
		}

		return nil
	}

	if err := ctx.ReplyEphemeralComplex(
		&interactions.InteractionCallbackData{
			Embeds:     []messages.Embed{data.embed},
			Components: []components.Component{},
		},
	); err != nil {
		responder.logger.Error(
			"failed to send interaction response",
			slog.String("error", err.Error()),
		)

		return fmt.Errorf(
			"responder: failed to send interaction response: %w",
			err,
		)
	}

	return nil
}
