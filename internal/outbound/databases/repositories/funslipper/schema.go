package funslipper

import (
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/uptrace/bun"
)

type SchemaFunSlipper struct {
	bun.BaseModel `bun:"table:fun_slippers"`
	ID            int `bun:"id,pk,autoincrement"`

	GuildID string `bun:"guild_id,notnull,unique:composite"`
	UserID  string `bun:"user_id,notnull,unique:composite"`
	Count   int    `bun:"count,notnull,default:1"`
}

func NewSchemaFunSplipper(funslipper domain.FunSlipper) *SchemaFunSlipper {
	schema := SchemaFunSlipper{
		ID:      funslipper.ID,
		GuildID: funslipper.GuildID,
		UserID:  funslipper.UserID,
		Count:   funslipper.Count,
	}

	return &schema
}

func (schema *SchemaFunSlipper) NewDomainFunSlipper() *domain.FunSlipper {
	domain := domain.FunSlipper{
		ID:      schema.ID,
		GuildID: schema.GuildID,
		UserID:  schema.UserID,
		Count:   schema.Count,
	}

	return &domain
}
