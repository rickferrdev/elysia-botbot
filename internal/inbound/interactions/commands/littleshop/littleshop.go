package littleshop

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/users"
	"github.com/google/uuid"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/interactionflow"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/littleshop"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
	"github.com/rickferrdev/elysia-botbot/internal/inbound/interactions/resembeds"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type (
	FxArgs struct {
		fx.In

		DiscordClient   *discordclient.DiscordClient
		Responder       responder.Interface
		InteractionFlow interactionflow.Interface
		LittleShop      littleshop.Interface
	}

	Command struct {
		client     *discordclient.DiscordClient
		responder  responder.Interface
		flow       interactionflow.Interface
		littleshop littleshop.Interface
	}
)

func New(args FxArgs) *Command {
	cmd := Command{
		client:     args.DiscordClient,
		responder:  args.Responder,
		flow:       args.InteractionFlow,
		littleshop: args.LittleShop,
	}

	cmd.client.Router.Command(
		"littleshop",
		"Lojinha da Elysia para compra de itens.",
		cmd.execute,
	)

	return &cmd
}

func (cmd *Command) execute(ctx *bot.InteractionContext) {
	cmd.ShopSetup(ctx)
}

func (cmd *Command) ShopSetup(ctx *bot.InteractionContext) {
	if err := ctx.DeferEphemeral(); err != nil {
		return
	}

	embed := resembeds.EmbedStartupShop(ctx.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{}))
	cmd.responder.Sender(ctx, responder.WithEmbed(embed), nil)

	menuID := uuid.NewString()
	interaction, err := cmd.flow.AwaitStringSelect(
		ctx,
		[]components.SelectOption{
			components.NewSelectOptionBuilder().
				SetLabel("Tomates").
				SetEmoji(map[string]string{"name": "🍅"}).
				SetValue("tomatoes").
				SetDescription(fmt.Sprintf("Preço: %d chinelos cada", domain.TomatoesCost)).
				Build(),
		},
		components.NewStringSelectMenuBuilder().
			SetCustomID(menuID).
			SetPlaceholder("Selecione um item para comprar.").
			SetMinValues(1).
			SetMaxValues(1).
			Build(),
	)
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	switch interaction.Values()[0] {
	case "tomatoes":
		cmd.PurchaseTomatoes(interaction)
	default:
		return
	}
}

func (cmd *Command) PurchaseTomatoes(ctx *bot.InteractionContext) {
	inputID := uuid.NewString()
	modalID := uuid.NewString()

	interaction, err := cmd.flow.AwaitSimpleTextModal(
		ctx,
		components.NewTextInputBuilder().
			SetCustomID(inputID).
			SetLabel("Quantidade a comprar").
			SetStyle(components.TextInputStyleShort).
			SetPlaceholder("Ex.: 1, 5, 10").
			SetRequired(true).
			SetMinLength(1).
			SetMaxLength(5).
			Build(),
		components.NewModalBuilder().
			SetCustomID(modalID).
			SetTitle("Comprar Tomates 🍅").
			Build(),
	)
	if err != nil {
		cmd.responder.Sender(ctx, responder.WithInternal, err)
		return
	}

	if err := interaction.DeferUpdate(); err != nil {
		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	quantity, err := strconv.Atoi(strings.TrimSpace(interaction.ModalValue(inputID)))
	if err != nil || quantity <= 0 {
		cmd.responder.Sender(interaction, responder.WithBadArguments, errors.New("quantidade inválida informada"))
		return
	}

	guildID := interaction.GuildID().String()
	if interaction.Guild != nil {
		guildID = interaction.Guild.ID.String()
	}
	userID := interaction.User().ID.String()

	balance, err := cmd.littleshop.PurchaseTomato(interaction.Context(), guildID, userID, quantity)
	if err != nil {
		if errors.Is(err, littleshop.InsufficientBalanceError) {
			embed := resembeds.EmbedInsufficientBalance(interaction.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{}))
			cmd.responder.Sender(interaction, responder.WithEmbed(embed), err)
			return
		}

		cmd.responder.Sender(interaction, responder.WithInternal, err)
		return
	}

	cost := quantity * domain.TomatoesCost
	embed := resembeds.EmbedPurchaseShop(interaction.Bot.User().DisplayAvatarURL(users.AvatarURLOptions{}), quantity, cost, balance)
	cmd.responder.Sender(interaction, responder.WithEmbed(embed), nil)
}
