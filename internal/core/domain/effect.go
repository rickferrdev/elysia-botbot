package domain

import "time"

type (
	EffectType  string
	EffectScope string
)

var (
	EffectTypeKwaii     EffectType = "kwaii"
	EffectTypeChristmas EffectType = "christmas"
)

type Effect struct {
	ID         int
	Type       EffectType
	Scope      EffectScope
	UserID     string
	GuildID    string
	Expiration time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
