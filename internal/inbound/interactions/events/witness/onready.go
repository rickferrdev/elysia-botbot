package witness

import (
	"context"
	"log/slog"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/rickferrdev/elysia-botbot/internal/infra/discordclient"
	"go.uber.org/fx"
)

var Invoke = fx.Invoke(New)

type FxArgs struct {
	fx.In

	Logger        *slog.Logger
	DiscordClient *discordclient.DiscordClient
}

type Event struct {
	client *discordclient.DiscordClient
	logger *slog.Logger
}

func New(args FxArgs) {
	event := Event{
		logger: args.Logger,
		client: args.DiscordClient,
	}

	args.DiscordClient.Bot.OnReady(func(ctx *bot.ReadyContext) {
		event.OnReady(ctx)
	})
}

func (event *Event) OnReady(ctx *bot.ReadyContext) {
	activities := []string{
		"As pessoas realmente leem isso?",
		"Comendo um pão de queijo",
		"Resolvendo o bug #12",
		"Levando meu dono para passear",
		"Robert Griesemer, Rob Pike e Ken Thompson criaram o Go no Google em 2007.",
		"Go oferece 2 recursos capazes de substituir a herança de Classe. A primeira é o embedding e a outra são as interfaces.",
	}

	ticker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()

	update := func(activity string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := event.client.Bot.SetActivity(ctx, bot.Activity{
			Name: activity,
			Type: 0,
		}); err != nil {
			event.logger.Error("failed to set activity", "error", err)
		}
	}

	index := 0

	update(activities[index])

	for range ticker.C {
		index = (index + 1) % len(activities)
		update(activities[index])
	}
}
