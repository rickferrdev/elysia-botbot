package responder

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type Message struct {
	Title            string
	Description      string
	WithTimestampNow bool
	WithColor        bool
}

type Interface interface {
	SendError(ctx *bot.InteractionContext, err error)
	SendMessage(ctx *bot.InteractionContext, message Message)
}

type Responder struct {
	logger *slog.Logger
}

type Params struct {
	fx.In
	Logger *slog.Logger
}

func New(params Params) *Responder {
	responder := Responder{
		logger: params.Logger,
	}
	return &responder
}

func (responder Responder) SendError(ctx *bot.InteractionContext, err error) {
	var builder strings.Builder
	builder.WriteString("```\n")
	builder.WriteString(err.Error())
	builder.WriteString("```\n")

	embed := messages.NewEmbedBuilder().
		SetTitle("𖥻 ׁ ׅ  An exception occurred.... ! ﹒  ◡◡").
		SetDescription(builder.String()).
		SetTimestampNow().
		SetColor(0xfdfd96)

	responder.handler(ctx, err, embed.Build())
}

func (responder Responder) SendMessage(ctx *bot.InteractionContext, message Message) {
	embed := messages.NewEmbedBuilder()
	if message.Title != "" {
		embed.SetTitle(fmt.Sprintf("𖥻 ׁ ׅ  %v ! ﹒  ◡◡", message.Title))
	}
	if message.Description != "" {
		embed.SetDescription(fmt.Sprintf("⋆˚꩜｡ %v", message.Description))
	}
	if message.WithColor {
		embed.SetColor(0xfdfd96)
	}
	if message.WithTimestampNow {
		embed.SetTimestampNow()
	}

	responder.handler(ctx, nil, embed.Build())
}

func (responder Responder) handler(ctx *bot.InteractionContext, err error, embed messages.Embed) {
	if err != nil {
		responder.logger.Error(err.Error())
	}
	if ctx == nil || !ctx.IsRepliable() {
		return
	}

	if ctx.Replied() {
		_, serr := ctx.FollowupComplex(rest.ExecuteWebhookParams{
			Embeds: []messages.Embed{embed},
			Flags:  messages.FlagEphemeral,
		})
		if serr != nil {
			responder.logger.Error(serr.Error())
		}
		return
	}

	if ctx.Deferred() {
		_, serr := ctx.EditReplyComplex(rest.EditMessageParams{
			Embeds: &[]messages.Embed{embed},
		})
		if serr != nil {
			responder.logger.Error(serr.Error())
		}
		return
	}

	if serr := ctx.ReplyEphemeralComplex(&interactions.InteractionCallbackData{
		Embeds: []messages.Embed{embed},
	}); serr != nil {
		responder.logger.Error(serr.Error())
	}
}
