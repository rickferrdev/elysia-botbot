package env

import (
	"github.com/rickferrdev/dotenv"
	"go.uber.org/fx"
)

var Provide = fx.Provide(New)

type Env struct {
	DiscordToken string `env:"DISCORD_TOKEN" required:"true"`
	GuildID      string `env:"GUILD_ID" required:"true"`
	GeminiAPIKey string `env:"GEMINI_API_KEY" required:"true"`
	RedisHost    string `env:"REDIS_HOST" required:"true"`
	RedisUser    string `env:"REDIS_USER" required:"true"`
	RedisPass    string `env:"REDIS_PASS" required:"true"`
	RedisDB      string `env:"REDIS_DB" required:"true"`
	SqliteDB     string `env:"SQLITE_DB" required:"true"`
}

func New() (*Env, error) {
	var env Env
	dotenv.Collect()

	if err := dotenv.Unmarshal(&env); err != nil {
		return nil, err
	}

	return &env, nil
}
