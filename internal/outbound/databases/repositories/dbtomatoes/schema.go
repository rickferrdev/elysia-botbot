package dbtomatoes

import (
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/uptrace/bun"
)

type SchemaTomatoes struct {
	bun.BaseModel `bun:"table:tomatoes"`

	ID        int    `bun:"id,pk,autoincrement"`
	GuildID   string `bun:"guild_id,notnull,unique:composite"`
	UserID    string `bun:"user_id,notnull,unique:composite"`
	Inventory int    `bun:"inventory,notnull,default:0"`
}

func NewSchemaTomatoes(tomatoes domain.Tomatoes) *SchemaTomatoes {
	schema := SchemaTomatoes{
		ID:        tomatoes.ID,
		GuildID:   tomatoes.GuildID,
		UserID:    tomatoes.UserID,
		Inventory: tomatoes.Inventory,
	}

	return &schema
}

func (schema *SchemaTomatoes) NewDomainTomatoes() *domain.Tomatoes {
	domain := domain.Tomatoes{
		ID:        schema.ID,
		GuildID:   schema.GuildID,
		UserID:    schema.UserID,
		Inventory: schema.Inventory,
	}

	return &domain
}
