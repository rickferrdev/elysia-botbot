package responder

import (
	"log/slog"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct {
		Logger *slog.Logger
	}

	FxArgs struct {
		fx.In
		Logger *slog.Logger
	}

	Interface interface {
		Sender(ctx *bot.InteractionContext, responder interactions.InteractionCallbackData, cause error)
	}

	ResponderCode    int
	ResponderContent string
	ResponderColor   int
)

const (
	CodeOK ResponderCode = iota + 1
	CodeCreated
	CodeNoContent
	CodeBadArguments
	CodeUnauthorized
	CodeForbidden
	CodeNotFound
	CodeConflict
	CodeInternal
	CodeServiceUnavailable
)

type Definition struct {
	Title       string
	Description string
	Color       ResponderColor
}

const (
	ColorSuccess ResponderColor = 0xA6E3A1
	ColorError   ResponderColor = 0xF38BA8
	ColorInfo    ResponderColor = 0x89B4FA
	ColorNeutral ResponderColor = 0xCDD6F4
)

var embeds = map[ResponderCode]messages.Embed{
	CodeOK: {
		Title:       "♡ Tudo certinho!",
		Description: "꒰ Sua solicitação foi concluída com sucesso. ✧ ꒱",
		Color:       int(ColorSuccess),
	},
	CodeCreated: {
		Title:       "୨୧ Criado com sucesso!",
		Description: "꒰ Prontinho! Seu recurso já está disponível. ♡ ꒱",
		Color:       int(ColorSuccess),
	},
	CodeNoContent: {
		Title:       "✧ Tudo certo!",
		Description: "꒰ Sua solicitação foi concluída. ♡ ꒱",
		Color:       int(ColorNeutral),
	},
	CodeBadArguments: {
		Title:       "૮꒰ ˶• ༝ •˶꒱ა Ops, vamos conferir?",
		Description: "♡ Algumas informações precisam ser verificadas. Confira os dados e tente novamente, tá bom?",
		Color:       int(ColorError),
	},
	CodeUnauthorized: {
		Title:       "୨୧ Precisamos confirmar sua identidade",
		Description: "꒰ Faça sua autenticação antes de continuar. ♡ ꒱",
		Color:       int(ColorError),
	},
	CodeForbidden: {
		Title:       "૮꒰ ˶>⤙<˶ ꒱ა Acesso não permitido",
		Description: "꒰ Você não tem permissão para realizar essa ação. ♡ ꒱",
		Color:       int(ColorError),
	},
	CodeNotFound: {
		Title:       "୨୧ Não encontramos isso...",
		Description: "꒰ O item solicitado não foi encontrado. Confira os dados, tá bom? ♡ ꒱",
		Color:       int(ColorInfo),
	},
	CodeConflict: {
		Title:       "✧ Ops! Temos um pequeno conflito",
		Description: "꒰ Essa ação não pode ser concluída agora. Confira os dados e tente novamente. ♡ ꒱",
		Color:       int(ColorError),
	},
	CodeInternal: {
		Title:       "૮꒰ ˶• ༝ •˶꒱ა Algo deu errado...",
		Description: "꒰ Não conseguimos concluir sua solicitação agora. Tente novamente daqui a pouquinho. ♡ ꒱",
		Color:       int(ColorError),
	},
	CodeServiceUnavailable: {
		Title:       "✧ Voltamos já, tá bom?",
		Description: "꒰ O serviço está temporariamente indisponível. Tente novamente em instantes. ♡ ꒱",
		Color:       int(ColorError),
	},
}

var sends = map[ResponderCode]interactions.InteractionCallbackData{
	CodeOK: {
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embeds[CodeOK]},
	},
	CodeCreated: {
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embeds[CodeCreated]},
	},
	CodeNoContent: {
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embeds[CodeNoContent]},
	},
	CodeBadArguments: {
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embeds[CodeBadArguments]},
	},
	CodeUnauthorized: {
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embeds[CodeUnauthorized]},
	},
	CodeForbidden: {
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embeds[CodeForbidden]},
	},
	CodeNotFound: {
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embeds[CodeNotFound]},
	},
	CodeConflict: {
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embeds[CodeConflict]},
	},
	CodeInternal: {
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embeds[CodeInternal]},
	},
	CodeServiceUnavailable: {
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embeds[CodeServiceUnavailable]},
	},
}

var WithOk = sends[CodeOK]
var WithCreated = sends[CodeCreated]
var WithNoContent = sends[CodeNoContent]
var WithBadArguments = sends[CodeBadArguments]
var WithUnauthorized = sends[CodeUnauthorized]
var WithForbidden = sends[CodeForbidden]
var WithNotFound = sends[CodeNotFound]
var WithConflict = sends[CodeConflict]
var WithInternal = sends[CodeInternal]
var WithServiceUnavailable = sends[CodeServiceUnavailable]

func New(args FxArgs) *Service {
	service := Service{Logger: args.Logger}
	return &service
}

func WithEmbed(embed messages.Embed) interactions.InteractionCallbackData {
	return interactions.InteractionCallbackData{
		Content: "‎. ݁⋆ ۶ৎ ݁˖ . ݁",
		Embeds:  []messages.Embed{embed},
	}
}

func (service *Service) Sender(ctx *bot.InteractionContext, responder interactions.InteractionCallbackData, cause error) {
	if cause != nil {
		slog.Error("response cause", slog.String("cause", cause.Error()))
	}
	if ctx == nil {
		service.Logger.Error("cannot send interaction response: context is nil")
		return
	}

	for index, values := range responder.Embeds {
		if err := values.Validate(); err != nil {
			responder.Embeds[index] = messages.Embed{}
			service.Logger.ErrorContext(
				ctx.Context(),
				"invalid embed in response",
				slog.Int("index", index), slog.String("cause", cause.Error()),
			)
			return
		}

		// used in all embeds
		values.Timestamp = time.Now().Format(time.RFC3339)
	}

	switch {
	case ctx.Deferred(), ctx.Replied():
		if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
			Content:     &responder.Content,
			Embeds:      &responder.Embeds,
			Components:  &responder.Components,
			Attachments: &ctx.Message.Attachments,
			Flags:       &responder.Flags,
		}); err != nil {
			service.Logger.Error("failed to edit interaction reply", slog.String("error", err.Error()))
		}
	default:
		if err := ctx.ReplyComplex(&responder); err != nil {
			service.Logger.Error("failed to send initial interaction reply", slog.String("error", err.Error()))
		}
	}
}
