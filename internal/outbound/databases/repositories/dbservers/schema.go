package dbservers

import (
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/uptrace/bun"
)

type SchemaServer struct {
	bun.BaseModel `bun:"table:servers,alias:server"`

	ID        int           `bun:"id,pk,autoincrement"`
	GuildID   string        `bun:"guild_id,notnull,unique"`
	Config    *SchemaConfig `bun:"rel:has-one,join:guild_id=guild_id"`
	CreatedAt time.Time     `bun:"created_at"`
	UpdatedAt time.Time     `bun:"updated_at"`
}

type SchemaConfig struct {
	bun.BaseModel `bun:"table:configs,alias:config"`

	ID        int           `bun:"id,pk,autoincrement"`
	GuildID   string        `bun:"guild_id,notnull,unique"`
	Prefix    string        `bun:"prefix,notnull"`
	Server    *SchemaServer `bun:"rel:belongs-to,join:guild_id=guild_id"`
	CreatedAt time.Time     `bun:"created_at"`
	UpdatedAt time.Time     `bun:"updated_at"`
}

func NewSchemaServer(server *domain.Server) *SchemaServer {
	if server == nil {
		return nil
	}

	if server.Config == nil {
		return nil
	}

	schema := SchemaServer{
		ID:        server.ID,
		GuildID:   server.GuildID,
		Config:    NewSchemaConfig(server.Config),
		CreatedAt: server.CreatedAt,
		UpdatedAt: server.UpdatedAt,
	}

	return &schema
}

func (schema *SchemaServer) NewDomainServer() *domain.Server {
	domain := domain.Server{
		GuildID:   schema.GuildID,
		Config:    schema.Config.NewDomainConfig(),
		CreatedAt: schema.CreatedAt,
		UpdatedAt: schema.UpdatedAt,
	}

	return &domain
}

func NewSchemaConfig(config *domain.Config) *SchemaConfig {
	if config == nil {
		return nil
	}

	schema := SchemaConfig{
		ID:        config.ID,
		GuildID:   config.GuildID,
		Prefix:    config.Prefix,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}

	return &schema
}

func (schema *SchemaConfig) NewDomainConfig() *domain.Config {
	if schema == nil {
		return nil
	}

	domain := domain.Config{
		GuildID:   schema.GuildID,
		ID:        schema.ID,
		Prefix:    schema.Prefix,
		CreatedAt: schema.CreatedAt,
		UpdatedAt: schema.UpdatedAt,
	}

	return &domain
}
