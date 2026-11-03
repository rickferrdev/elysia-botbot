package tomato

import (
	"errors"
	"fmt"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/users"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/tomatoes"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/resembeds"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type (
	FxArgs struct {
		fx.In

		TomatoesService tomatoes.Interface
		DiscordClient   *discordclient.DiscordClient
		Responder       responder.Interface
	}

	Command struct {
		tomatoes  tomatoes.Interface
		client    *discordclient.DiscordClient
		responder responder.Interface
	}
)

func New(args FxArgs) *Command {
	cmd := Command{
		tomatoes:  args.TomatoesService,
		client:    args.DiscordClient,
		responder: args.Responder,
	}

	cmd.client.Router.Command(
		"tomato",
		"Gaste 1 🍅 para jogar em alguém; cada tomate custa 5 🩴.",
		cmd.execute,
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeUser,
			Name:        "user",
			Description: "Quem vai levar a tomatada?",
		},
	)
	cmd.client.Router.Prefix("tomato", cmd.executeInMessage).
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
	avatar := ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})

	balance, err := cmd.tomatoes.ThrowTomato(ctx.Context(), ctx.GuildID().String(), author)
	if err != nil {
		if errors.Is(err, tomatoes.InsufficientBalanceError) {
			cmd.responder.Sender(ctx, responder.WithEmbed(resembeds.EmbedInsufficientBalance(avatar)), err)
			return
		}

		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	var content string
	switch {
	case author != "" && destination == "0" || author == destination:
		content = fmt.Sprintf("***<@%s> gastou 1 🍅 para levar uma tomatada de si mesmo! saldo: %d 🍅.***", author, balance)
		if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
			Content: &content,
			Embeds:  &[]messages.Embed{resembeds.EmbedThrowTomato()},
		}); err != nil {
			cmd.responder.Sender(ctx, responder.WithInternal, err)
			return
		}
	case destination != "0":
		content = fmt.Sprintf("***<@%s> levou uma tomatada de <@%s>! saldo de <@%s>: %d 🍅.***", destination, author, author, balance)
		if _, err := ctx.EditReplyComplex(rest.EditMessageParams{
			Content: &content,
			Embeds:  &[]messages.Embed{resembeds.EmbedThrowTomato()},
		}); err != nil {
			cmd.responder.Sender(ctx, responder.WithInternal, err)
			return
		}
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
	avatar := ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{})

	balance, err := cmd.tomatoes.ThrowTomato(ctx.Context(), ctx.GuildID.String(), author)
	if err != nil {
		if errors.Is(err, tomatoes.InsufficientBalanceError) {
			cmd.responder.SenderContent(ctx, responder.Content{
				Title:       resembeds.EmbedInsufficientBalance(avatar).Title,
				Description: resembeds.EmbedInsufficientBalance(avatar).Description,
			}, err)
			return
		}

		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
		return
	}

	if err := cmd.client.Bot.Rest.DeleteMessage(ctx.Context(), channelID, messageID); err != nil {
		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
		return
	}

	var content string

	switch {
	case author != "" && destination == "" || author == destination:
		content = fmt.Sprintf("***<@%s> gastou 1 🍅 para levar uma tomatada de si mesmo! Saldo: %d 🍅.***", author, balance)
		if _, err := cmd.client.Bot.Rest.CreateMessageComplex(ctx.Context(), channelID, messages.MessageSend{
			Content: content,
			Embeds:  []messages.Embed{resembeds.EmbedThrowTomato()},
		}); err != nil {
			cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
			return
		}
		return
	case destination != "":
		content = fmt.Sprintf("***%s levou uma tomatada de <@%s>! Saldo de <@%s>: %d 🍅.***", destination, author, author, balance)
		if _, err := cmd.client.Bot.Rest.CreateMessageComplex(ctx.Context(), channelID, messages.MessageSend{
			Content: content,
			Embeds:  []messages.Embed{resembeds.EmbedThrowTomato()},
		}); err != nil {
			cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
			return
		}
		return
	}
}

func (cmd *Command) aliases() []string {
	return []string{"t"}
}
