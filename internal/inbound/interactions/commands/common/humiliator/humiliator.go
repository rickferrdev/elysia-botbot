package humiliator

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/users"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

const maxReasonLength = 200

var (
	errEmptyReason = errors.New("o motivo não pode ficar vazio")
	errLongReason  = errors.New("o motivo deve ter no máximo 200 caracteres")
)

type (
	FxArgs struct {
		fx.In

		DiscordClient *discordclient.DiscordClient
		Responder     responder.Interface
	}

	Command struct {
		client    *discordclient.DiscordClient
		responder responder.Interface
	}
)

func New(args FxArgs) *Command {
	cmd := Command{
		client:    args.DiscordClient,
		responder: args.Responder,
	}

	cmd.client.Router.Command(
		"humiliator",
		"Registre uma coisa errada que alguém fez.",
		cmd.execute,
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeUser,
			Name:        "member",
			Description: "Quem fez a coisa boba?",
			Required:    true,
		},
		interactions.ApplicationCommandOption{
			Type:        interactions.ApplicationCommandOptionTypeString,
			Name:        "reason",
			Description: "O que essa pessoa fez?",
			Required:    true,
			MaxLength:   200,
		},
	).RequireBotPermissions(permissions.AddReactions)
	cmd.client.Router.Prefix("humiliator", cmd.executeInMessage).
		Aliases(cmd.aliases()...).
		Use(bot.PrefixCooldown(time.Second * 2))

	return &cmd
}

func (cmd *Command) executeInMessage(ctx *bot.MessageContext, args []string) {
	if len(ctx.Message.Mentions) == 0 {
		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeBadArguments), fmt.Errorf("você precisa mencionar alguém"))
		return
	}
	memberID := ctx.Message.Mentions[0].ID
	reason := strings.TrimSpace(strings.Join(args[1:], " "))
	memberAvatar := ctx.Message.Mentions[0].AvatarURL(users.AvatarURLOptions{})
	embed, err := buildEmbed(memberID.String(), memberAvatar, reason)
	if err != nil {
		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeBadArguments), err)
		return
	}

	message, err := cmd.client.Bot.Rest.CreateMessageComplex(ctx.Context(), ctx.ChannelID(), messages.MessageSend{
		Embeds: []messages.Embed{embed},
	})
	if err != nil {
		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
		return
	}

	for _, emoji := range []string{"🍅", "🩴"} {
		if err := cmd.client.Bot.Rest.CreateReaction(ctx.Context(), ctx.ChannelID(), message.ID, emoji); err != nil {
			cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
			return
		}
	}

	if err := ctx.Bot.Rest.DeleteMessage(ctx.Context(), ctx.ChannelID(), ctx.ID); err != nil {
		cmd.responder.SenderContent(ctx, responder.WithCodeContent(responder.CodeInternal), err)
	}
}

func (cmd *Command) execute(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	memberID := ctx.OptionUser("member")
	reason := strings.TrimSpace(ctx.OptionString("reason"))
	user, err := ctx.Bot.Rest.GetUser(ctx.Context(), memberID)
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}
	embed, err := buildEmbed(memberID.String(), user.AvatarURL(users.AvatarURLOptions{}), reason)
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithBadArguments, err)
		return
	}

	message, err := cmd.client.Bot.Rest.CreateMessageComplex(ctx.Context(), ctx.ChannelID(), messages.MessageSend{
		Embeds: []messages.Embed{embed},
	})
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	for _, emoji := range []string{"🍅", "🩴"} {
		if err := cmd.client.Bot.Rest.CreateReaction(ctx.Context(), ctx.ChannelID(), message.ID, emoji); err != nil {
			cmd.responder.Sender(ctx, responder.WithInternal, err)
			return
		}
	}

	if err := ctx.DeleteReply(); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
	}
}

func buildEmbed(memberID, avatarURL, reason string) (messages.Embed, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return messages.Embed{}, errEmptyReason
	}
	if utf8.RuneCountInString(reason) > maxReasonLength {
		return messages.Embed{}, errLongReason
	}

	builder := messages.NewEmbedBuilder().
		SetTitle(fmt.Sprintf("Membro bobo! Fez coisa errada: %s", reason)).
		SetDescription(fmt.Sprintf("Pessoa citada: <@%s>", memberID)).
		SetColor(0xF38BA8).
		SetTimestampNow()
	if avatarURL != "" {
		builder.SetImage(avatarURL)
	}

	return builder.Build(), nil
}

func (cmd *Command) aliases() []string {
	return []string{"h"}
}
