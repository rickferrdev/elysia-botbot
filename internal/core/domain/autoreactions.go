package domain

type AutoReactionType string

var (
	AutoReactionTypeAttachment AutoReactionType = "attachment"
	AutoReactionTypeMessage    AutoReactionType = "message"
)

type AutoReactions struct {
	ID        int
	GuildID   string
	ChannelID string
	Type      AutoReactionType
	Emoji     string
}
