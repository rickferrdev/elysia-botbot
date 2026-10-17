package domain

type Server struct {
	ID      int
	GuildID string
	Config  *Config
}

type Config struct {
	ID      int
	GuildID string
	Prefix  string
	Server  *Server
}
