package bestchoice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/james-see/swatchify/pkg/swatchify"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
	_ "golang.org/x/image/webp"
)

var Invoke = fx.Invoke(New)

type FxArgs struct {
	fx.In

	Responder       responder.Interface
	InteractionFlow interactionflow.Interface
	DiscordClient   *discordclient.DiscordClient
}

type Command struct {
	flow      interactionflow.Interface
	client    *discordclient.DiscordClient
	responder responder.Interface
}

func New(args FxArgs) *Command {
	cmd := Command{
		flow:      args.InteractionFlow,
		client:    args.DiscordClient,
		responder: args.Responder,
	}

	cmd.client.Router.Command("bestchoice", "Qual imagem é melhor?", cmd.execute,
		interactions.ApplicationCommandOption{
			Name:        "one-title",
			Description: "Qual título da primeira foto?",
			Required:    true,
			Type:        interactions.ApplicationCommandOptionTypeString,
		},
		interactions.ApplicationCommandOption{
			Name:        "one",
			Description: "Selecione uma foto",
			Required:    true,
			Type:        interactions.ApplicationCommandOptionTypeAttachment,
		},
		interactions.ApplicationCommandOption{
			Name:        "two-title",
			Description: "Qual título da segunda foto?",
			Required:    true,
			Type:        interactions.ApplicationCommandOptionTypeString,
		},
		interactions.ApplicationCommandOption{
			Name:        "two",
			Description: "Selecione uma foto",
			Required:    true,
			Type:        interactions.ApplicationCommandOptionTypeAttachment,
		},
	).RequirePermissions(permissions.Administrator).RequireBotPermissions(permissions.AttachFiles)

	return &cmd
}

type resolved struct {
	Resolved struct {
		Attachments map[string]struct {
			ID          string `json:"id"`
			Filename    string `json:"filename"`
			URL         string `json:"url"`
			ProxyURL    string `json:"proxy_url"`
			ContentType string `json:"content_type"`
			Size        int    `json:"size"`
		} `json:"attachments"`
	} `json:"resolved"`
}

const maxImageSize = 10 << 20 // 10 MB

func (cmd *Command) execute(ctx *bot.InteractionContext) {
	if err := ctx.Defer(); err != nil {
		return
	}

	image := ctx.Option("one")
	if image.Type != interactions.ApplicationCommandOptionTypeAttachment {
		return
	}

	title := ctx.OptionString("one-title")
	url, color, err := cmd.helperObtainURL(ctx, image.Value.(string))
	var choices []messages.Embed
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithBadArguments, err)
		return
	}
	choices = append(choices, messages.NewEmbedBuilder().SetTitle(title).SetImage(url).SetColorHex(color).SetTimestampNow().Build())

	image = ctx.Option("two")
	if image.Type != interactions.ApplicationCommandOptionTypeAttachment {
		return
	}

	url, color, err = cmd.helperObtainURL(ctx, image.Value.(string))
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithBadArguments, err)
		return
	}
	title = ctx.OptionString("two-title")
	choices = append(choices, messages.NewEmbedBuilder().SetTitle(title).SetImage(url).SetColorHex(color).SetTimestampNow().Build())

	message, err := cmd.client.Bot.Rest.CreateMessageComplex(ctx.Context(), ctx.ChannelID(), messages.MessageSend{
		Embeds: []messages.Embed{choices[0]},
	})
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}
	if err := cmd.client.Bot.Rest.CreateReaction(ctx.Context(), ctx.ChannelID(), message.ID, "💖"); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	message, err = cmd.client.Bot.Rest.CreateMessageComplex(ctx.Context(), ctx.ChannelID(), messages.MessageSend{
		Embeds: []messages.Embed{choices[1]},
	})
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}
	if err := cmd.client.Bot.Rest.CreateReaction(ctx.Context(), ctx.ChannelID(), message.ID, "💖"); err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}
	ctx.DeleteReply()
}

func (cmd *Command) helperObtainURL(ctx *bot.InteractionContext, id string) (string, string, error) {
	raw, err := json.Marshal(ctx.Interaction.Data)
	if err != nil {
		return "", "", err
	}

	var data resolved
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", "", err
	}

	att, ok := data.Resolved.Attachments[id]
	if !ok {
		return "", "", nil
	}

	color, err := cmd.swatchesFromURL(ctx.Context(), att.URL)
	if err != nil {
		return "", "", err
	}

	return att.URL, color[0].Hex, nil
}

func (cmd *Command) fetchImage(ctx context.Context, url string) (image.Image, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status inesperado: %s", resp.Status)
	}

	img, _, err := image.Decode(io.LimitReader(resp.Body, maxImageSize))
	return img, err
}

func (cmd *Command) swatchesFromURL(ctx context.Context, url string) ([]swatchify.Color, error) {
	img, err := cmd.fetchImage(ctx, url)
	if err != nil {
		return nil, err
	}

	return swatchify.ExtractFromImage(img, &swatchify.Options{NumColors: 1, Quality: 80})
}
