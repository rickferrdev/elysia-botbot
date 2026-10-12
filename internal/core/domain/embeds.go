package domain

import (
	"time"

	"github.com/discord-go/discord.go/messages"
	"github.com/uptrace/bun"
)

type Embed struct {
	bun.BaseModel
	ID int

	GuildID   string
	UserID    string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EmbedContentDecode struct {
	Content     string                `json:"content"`
	Embeds      []messages.Embed      `json:"embeds"`
	Attachments []messages.Attachment `json:"attachments"`
}
