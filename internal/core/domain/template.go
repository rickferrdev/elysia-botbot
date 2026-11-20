package domain

import (
	"time"

	"github.com/discord-go/discord.go/messages"
)

type Template struct {
	ID       int    `json:"id"`
	GuildID  string `json:"guild_id"`
	AuthorID string `json:"author_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	TemplateContent *TemplateContent `json:"template_content"`
}

type TemplateContent struct {
	ID         int `json:"id"`
	TemplateID int `json:"template_id"`

	Content     string `json:"content"`
	Attachments string `json:"attachments"`
	Embeds      string `json:"embeds"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Template *Template `json:"template"`
}

type EmbedContentDecode struct {
	Content     string                `json:"content"`
	Embeds      []messages.Embed      `json:"embeds"`
	Attachments []messages.Attachment `json:"attachments"`
}
