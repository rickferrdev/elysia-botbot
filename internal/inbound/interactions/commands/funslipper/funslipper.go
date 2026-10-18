package funslipper

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/discord-go/discord.go/bot"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/funslipper"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type FxArgs struct {
	fx.In

	Responder       responder.Interface
	FunSlipper      funslipper.Interface
	DiscordClient   *discordclient.DiscordClient
	InteractionFlow interactionflow.Interface
}

type Command struct {
	fx.In

	flow       interactionflow.Interface
	client     *discordclient.DiscordClient
	responder  responder.Interface
	funslipper funslipper.Interface
}

func New(args FxArgs) *Command {
	server := Command{
		responder:  args.Responder,
		funslipper: args.FunSlipper,
		client:     args.DiscordClient,
		flow:       args.InteractionFlow,
	}

	args.DiscordClient.Router.Prefix("slipper", server.execute).Aliases("s")
	return &server
}

func (cmd *Command) execute(ctx *bot.MessageContext, args []string) {
	cmd.FunSlipper(ctx, args)
}

func (cmd *Command) FunSlipper(ctx *bot.MessageContext, args []string) {
	authorID := ctx.Author.ID.String()
	userID := authorID
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		userID = normalizeUserID(strings.TrimSpace(strings.Join(args, " ")))
	}

	data, err := cmd.funslipper.IncrementCount(
		ctx.Context(),
		ctx.GuildID.String(),
		userID,
	)
	if err != nil {
		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
		return
	}

	message := fmt.Sprintf(
		"***[%dx]*** <@%s> foi ***atingido*** por uma chinelada do <@%s>!🩴",
		data.Count,
		userID,
		authorID,
	)

	if userID == authorID {
		message = fmt.Sprintf(
			"***[%dx]*** <@%s> deu uma chinelada em ***si mesmo***!🩴",
			data.Count,
			authorID,
		)
	}

	n := rand.Intn(10)
	if userID != authorID && n == 1 {
		message = fmt.Sprintf("***[Errou!]*** <@%s> foi ***antigido*** pela própria chinelada!🩴", authorID)
	}

	if err := ctx.Bot.Rest.DeleteMessage(
		ctx.Context(),
		ctx.ChannelID(),
		ctx.ID,
	); err != nil {
		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
		return
	}

	if _, err := ctx.Bot.Rest.CreateMessage(
		ctx.Context(),
		ctx.ChannelID(),
		message,
	); err != nil {
		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
		return
	}
}

func normalizeUserID(raw string) string {
	raw = strings.TrimPrefix(raw, "<@")
	raw = strings.TrimPrefix(raw, "!")
	return strings.TrimSuffix(raw, ">")
}
