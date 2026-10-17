package embeds

import (
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/uptrace/bun"
)

type SchemaEmbed struct {
	bun.BaseModel `bun:"table:embeds,alias:embed"`
	ID            int `bun:"id,pk,autoincrement"`

	GuildID   string    `bun:"guild_id,notnull"`
	AuthorID  string    `bun:"author_id,notnull"`
	Content   string    `bun:"content,notnull"`
	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`
}

func NewSchemaEmbeds(embed domain.Embed) *SchemaEmbed {
	schema := SchemaEmbed{
		ID:        embed.ID,
		GuildID:   embed.GuildID,
		AuthorID:  embed.UserID,
		Content:   embed.Content,
		CreatedAt: embed.CreatedAt,
		UpdatedAt: embed.UpdatedAt,
	}

	return &schema
}

func (schema *SchemaEmbed) NewDomainEmbed() *domain.Embed {
	domain := domain.Embed{
		ID:        schema.ID,
		GuildID:   schema.GuildID,
		UserID:    schema.AuthorID,
		Content:   schema.Content,
		CreatedAt: schema.CreatedAt,
		UpdatedAt: schema.UpdatedAt,
	}

	return &domain
}
