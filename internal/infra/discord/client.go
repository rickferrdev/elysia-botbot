package discord

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"go.uber.org/fx"
)

type DiscordClient struct {
	lifecycle fx.Lifecycle
	Bot       *bot.Bot
	Router    *bot.Router
	env       *env.Env
	logger    *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	guildID   snowflake.ID
}

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Env       *env.Env
	Logger    *slog.Logger
}

var Provide = fx.Provide(New)
var Invoke = fx.Invoke(func(discord *DiscordClient) error {
	return discord.Start()
})

func New(params Params) *DiscordClient {
	guildID := snowflake.MustParse(params.Env.GuildID)

	client := DiscordClient{
		env:       params.Env,
		lifecycle: params.Lifecycle,
		logger:    params.Logger,
		guildID:   guildID,
	}

	client.Router = bot.NewRouter()
	client.Bot = bot.New(
		params.Env.DiscordToken,
		bot.WithIntents(intents.Guilds),
		bot.WithRouter(client.Router),
		bot.WithGuildCommandSync(guildID),
	)

	return &client
}

func (discord *DiscordClient) Start() error {
	discord.lifecycle.Append(fx.Hook{
		OnStart: func(fctx context.Context) error {
			diagnosticCtx, cancel := context.WithTimeout(fctx, 10*time.Second)
			defer cancel()
			if err := discord.removeDuplicateGlobalCommands(diagnosticCtx); err != nil {
				return err
			}

			discord.ctx, discord.cancel = context.WithCancel(context.Background())
			go func() {
				if err := discord.Bot.RunContext(discord.ctx); err != nil {
					discord.logger.Error("error starting the bot", slog.Any("error", err))
				}
			}()

			return nil
		},
		OnStop: func(fctx context.Context) error {
			if discord.cancel != nil {
				discord.cancel()
			}
			return discord.Bot.Stop(fctx)
		},
	})

	return nil
}

func (discord *DiscordClient) removeDuplicateGlobalCommands(ctx context.Context) error {
	client := rest.New(discord.env.DiscordToken, nil, nil)

	var application struct {
		ID snowflake.ID `json:"id,string"`
	}
	if err := client.Request(ctx, "GET", "/applications/@me", nil, &application); err != nil {
		return fmt.Errorf("get Discord application: %w", err)
	}

	globalCommands, err := client.GetGlobalApplicationCommands(ctx, application.ID)
	if err != nil {
		return fmt.Errorf("list global Discord commands: %w", err)
	}

	localCommands := make(map[string]struct{})
	for _, command := range discord.Router.Commands() {
		localCommands[command.Name] = struct{}{}
	}

	for _, command := range globalCommands {
		if command.Type == nil || *command.Type != interactions.ApplicationCommandTypeChatInput {
			continue
		}
		if _, exists := localCommands[command.Name]; !exists {
			continue
		}

		discord.logger.Info(
			"removing duplicate global Discord command",
			slog.String("name", command.Name),
			slog.String("id", command.ID.String()),
		)
		if err := client.DeleteGlobalApplicationCommand(ctx, application.ID, command.ID); err != nil {
			return fmt.Errorf("delete duplicate global Discord command %q: %w", command.Name, err)
		}
	}

	return nil
}
