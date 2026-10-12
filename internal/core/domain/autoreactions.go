package domain

type AutoReactionType string

var (
	AutoReactionTypeAttachment AutoReactionType = "attachment"
	AutoReactionTypeMessage    AutoReactionType = "message"
)

type AutoReactions struct {
	ID        int              `bun:"id,pk,autoincrement"`
	GuildID   string           `bun:"guild_id,notnull,unique"`
	ChannelID string           `bun:"channel_id"`
	Type      AutoReactionType `bun:"type"`
	Emoji     string           `bun:"emoji"`
}
