package dbslipper

import (
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/uptrace/bun"
)

type SchemaSlipper struct {
	bun.BaseModel `bun:"table:slippers"`

	ID        int    `bun:"id,pk,autoincrement"`
	GuildID   string `bun:"guild_id,notnull,unique:composite"`
	UserID    string `bun:"user_id,notnull,unique:composite"`
	Inventory int    `bun:"inventory,notnull,default:0"`
}

func NewSchemaSlipper(slipper domain.Slipper) *SchemaSlipper {
	schema := SchemaSlipper{
		ID:        slipper.ID,
		GuildID:   slipper.GuildID,
		UserID:    slipper.UserID,
		Inventory: slipper.Inventory,
	}

	return &schema
}

func (schema *SchemaSlipper) NewDomainSlipper() *domain.Slipper {
	domain := domain.Slipper{
		ID:        schema.ID,
		GuildID:   schema.GuildID,
		UserID:    schema.UserID,
		Inventory: schema.Inventory,
	}

	return &domain
}
