package env

import (
	"github.com/rickferrdev/dotenv"
	"go.uber.org/fx"
)

var Provide = fx.Provide(New)

type Env struct {
	DiscordToken string `env:"DISCORD_TOKEN"`
	GuildID      string `env:"GUILD_ID"`
	RedisHost    string `env:"REDIS_HOST"`
	RedisPort    string `env:"REDIS_PORT"`
	RedisUser    string `env:"REDIS_USER"`
	RedisPass    string `env:"REDIS_PASS"`
	RedisDB      string `env:"REDIS_DB"`
	SqliteDB     string `env:"SQLITE_DB"`
}

func New() (*Env, error) {
	var env Env
	dotenv.Collect()

	if err := dotenv.Unmarshal(&env); err != nil {
		return nil, err
	}

	return &env, nil
}
