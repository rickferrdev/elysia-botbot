package domain

import "time"

type Server struct {
	ID        int
	GuildID   string
	Config    *Config
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Config struct {
	ID        int
	GuildID   string
	Prefix    string
	Server    *Server
	CreatedAt time.Time
	UpdatedAt time.Time
}
