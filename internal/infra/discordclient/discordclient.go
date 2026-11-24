package discordclient

import (
	"context"
	"log/slog"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/rickferrdev/elysia-botbot/internal/config/env"
	"github.com/rickferrdev/elysia-botbot/internal/infra/logger"
	"go.uber.org/fx"
)

type DiscordClient struct {
	Bot    *bot.Bot
	Router *bot.Router

	lifecycle fx.Lifecycle
	env       *env.Env
	logger    *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	guildID   snowflake.ID
}

type FxArgs struct {
	fx.In
	Lifecycle fx.Lifecycle
	Env       *env.Env
	Logger    *slog.Logger
}

var (
	Provide = fx.Provide(New)
	Invoke  = fx.Invoke(func(discord *DiscordClient) error {
		return discord.Start()
	})
)

func New(args FxArgs) *DiscordClient {
	guildID := snowflake.MustParse(args.Env.GuildID)

	client := DiscordClient{
		env:       args.Env,
		lifecycle: args.Lifecycle,
		logger:    args.Logger,
		guildID:   guildID,
	}

	client.Router = bot.NewRouter()
	client.Bot = bot.New(
		args.Env.DiscordToken,
		bot.WithPrefix("e!"),
		bot.WithIntents(intents.Guilds|intents.GuildMessages|intents.MessageContent),
		bot.WithRouter(client.Router),
		bot.WithGuildCommandSync(guildID),
	)

	return &client
}

func (discord *DiscordClient) Start() error {
	discord.lifecycle.Append(fx.Hook{
		OnStart: func(fctx context.Context) error {
			discord.ctx, discord.cancel = context.WithCancel(context.Background())
			go func() {
				if err := discord.Bot.RunContext(discord.ctx); err != nil {
					discord.logger.Error(logger.DiscordErrorOnStartup, slog.Any("error", err))
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
