package server

import (
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/uptrace/bun"
)

type SchemaServer struct {
	bun.BaseModel `bun:"table:servers,alias:server"`

	ID      int           `bun:"id,pk,autoincrement"`
	GuildID string        `bun:"guild_id,notnull,unique"`
	Config  *SchemaConfig `bun:"rel:has-one,join:guild_id=guild_id"`
}

type SchemaConfig struct {
	bun.BaseModel `bun:"table:configs,alias:config"`

	ID      int           `bun:"id,pk,autoincrement"`
	GuildID string        `bun:"guild_id,notnull,unique"`
	Prefix  string        `bun:"prefix,notnull"`
	Server  *SchemaServer `bun:"rel:belongs-to,join:guild_id=guild_id"`
}

func NewSchemaServer(server *domain.Server) *SchemaServer {
	if server == nil {
		return nil
	}

	schema := SchemaServer{
		ID:      server.ID,
		GuildID: server.GuildID,
		Config:  NewSchemaConfig(server.Config),
	}

	return &schema
}

func NewSchemaConfig(config *domain.Config) *SchemaConfig {
	if config == nil {
		return nil
	}

	schema := SchemaConfig{
		ID:      config.ID,
		GuildID: config.GuildID,
		Prefix:  config.Prefix,
	}

	return &schema
}
