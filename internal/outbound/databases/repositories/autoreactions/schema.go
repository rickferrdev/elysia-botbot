package autoreactions

import (
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/uptrace/bun"
)

type SchemaAutoReactions struct {
	bun.BaseModel `bun:"table:autoreactions,alias:autoreaction"`

	ID        int                     `bun:"id,pk,autoincrement"`
	GuildID   string                  `bun:"guild_id,notnull"`
	ChannelID string                  `bun:"channel_id,notnull,"`
	Type      domain.AutoReactionType `bun:"type,notnull"`
	Emoji     string                  `bun:"emoji,notnull"`
}

func NewSchemaAutoReactions(autoreactions domain.AutoReactions) *SchemaAutoReactions {
	schema := SchemaAutoReactions{
		ID:        autoreactions.ID,
		GuildID:   autoreactions.GuildID,
		ChannelID: autoreactions.ChannelID,
		Type:      autoreactions.Type,
		Emoji:     autoreactions.Emoji,
	}

	return &schema
}

func (schema *SchemaAutoReactions) NewDomainAutoReactions() *domain.AutoReactions {
	domain := domain.AutoReactions{
		ID:        schema.ID,
		GuildID:   schema.GuildID,
		ChannelID: schema.ChannelID,
		Type:      schema.Type,
		Emoji:     schema.Emoji,
	}

	return &domain
}
