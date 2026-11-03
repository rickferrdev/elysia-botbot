package slipper

import (
	"fmt"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/interactions"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/slipper"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type (
	FxArgs struct {
		fx.In

		SlipperService slipper.Interface
		DiscordClient  *discordclient.DiscordClient
		Responder      responder.Interface
	}

	Command struct {
		slipper   slipper.Interface
		client    *discordclient.DiscordClient
		responder responder.Interface
	}
)

func New(args FxArgs) *Command {
	cmd := Command{
		slipper:   args.SlipperService,
		client:    args.DiscordClient,
		responder: args.Responder,
	}

	cmd.client.Router.Command(
		"slipper",
		"Ganhe 1 🩴 (moeda) dando uma chinelada.",
		cmd.execute,
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeUser,
			Name:        "user",
			Description: "Quem vai levar a chinelada?",
		},
	)
	cmd.client.Router.Prefix("slipper", cmd.executeInMessage).
		Aliases(cmd.aliases()...).
		Use(bot.PrefixCooldown(time.Second * 2))

	return &cmd
}

func (cmd *Command) execute(ctx *bot.InteractionContext) {
	if err := ctx.Defer(); err != nil {
		return
	}

	destination := ctx.OptionUser("user").String()
	author := ctx.User().ID.String()

	inventory, err := cmd.slipper.ThrowSlipper(ctx.Context(), ctx.GuildID().String(), author)
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	var message string
	switch {
	case author != "" && destination == "0" || author == destination:
		message = fmt.Sprintf("***<@%s> levou uma chinelada de si mesmo! saldo: %d 🩴***", author, inventory)
		if err := ctx.EditReplyText(message); err != nil {
			cmd.responder.Sender(ctx, responder.WithInternal, err)
			return
		}
		return
	case destination != "0":
		message = fmt.Sprintf("***<@%s> levou uma chinelada de <@%s>! saldo: %d 🩴***", destination, author, inventory)
		if err := ctx.EditReplyText(message); err != nil {
			cmd.responder.Sender(ctx, responder.WithInternal, err)
			return
		}
		return
	}
}

func (cmd *Command) executeInMessage(ctx *bot.MessageContext, args []string) {
	destination, author := "", ""
	author = ctx.Author.ID.String()
	if len(args) > 0 && args[0] != "" {
		destination = args[0]
	}

	messageID := ctx.ID
	channelID := ctx.ChannelID()

	if err := cmd.client.Bot.Rest.DeleteMessage(ctx.Context(), channelID, messageID); err != nil {
		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
		return
	}

	inventory, err := cmd.slipper.ThrowSlipper(ctx.Context(), ctx.GuildID.String(), author)
	if err != nil {
		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
		return
	}

	var message string
	switch {
	case author != "" && destination == "" || author == destination:
		message = fmt.Sprintf("***<@%s> levou uma chinelada de si mesmo! saldo: %d 🩴***", author, inventory)
		if _, err := cmd.client.Bot.Rest.CreateMessage(ctx.Context(), channelID, message); err != nil {
			cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
			return
		}
		return
	case destination != "":
		message = fmt.Sprintf("***%s levou uma chinelada de <@%s>! saldo: %d 🩴***", destination, author, inventory)
		if _, err := cmd.client.Bot.Rest.CreateMessage(ctx.Context(), channelID, message); err != nil {
			cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
			return
		}
		return
	}
}

func (cmd *Command) aliases() []string {
	return []string{"s"}
}
