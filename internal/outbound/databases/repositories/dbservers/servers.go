package dbservers

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type SchemaServer struct {
	bun.BaseModel `bun:"table:servers,alias:server"`

	ID        int           `bun:"id,pk,autoincrement"`
	GuildID   string        `bun:"guild_id,notnull,unique"`
	Config    *SchemaConfig `bun:"rel:has-one,join:guild_id=guild_id"`
	CreatedAt time.Time     `bun:"created_at"`
	UpdatedAt time.Time     `bun:"updated_at"`
}

type SchemaConfig struct {
	bun.BaseModel `bun:"table:configs,alias:config"`

	ID        int           `bun:"id,pk,autoincrement"`
	GuildID   string        `bun:"guild_id,notnull,unique"`
	Prefix    string        `bun:"prefix,notnull"`
	Server    *SchemaServer `bun:"rel:belongs-to,join:guild_id=guild_id"`
	CreatedAt time.Time     `bun:"created_at"`
	UpdatedAt time.Time     `bun:"updated_at"`
}

func NewSchemaServer(server *domain.Server) *SchemaServer {
	if server == nil {
		return nil
	}

	if server.Config == nil {
		return nil
	}

	schema := SchemaServer{
		ID:        server.ID,
		GuildID:   server.GuildID,
		Config:    NewSchemaConfig(server.Config),
		CreatedAt: server.CreatedAt,
		UpdatedAt: server.UpdatedAt,
	}

	return &schema
}

func (schema *SchemaServer) NewDomainServer() *domain.Server {
	domain := domain.Server{
		GuildID:   schema.GuildID,
		Config:    schema.Config.NewDomainConfig(),
		CreatedAt: schema.CreatedAt,
		UpdatedAt: schema.UpdatedAt,
	}

	return &domain
}

func NewSchemaConfig(config *domain.Config) *SchemaConfig {
	if config == nil {
		return nil
	}

	schema := SchemaConfig{
		ID:        config.ID,
		GuildID:   config.GuildID,
		Prefix:    config.Prefix,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}

	return &schema
}

func (schema *SchemaConfig) NewDomainConfig() *domain.Config {
	if schema == nil {
		return nil
	}

	domain := domain.Config{
		GuildID:   schema.GuildID,
		ID:        schema.ID,
		Prefix:    schema.Prefix,
		CreatedAt: schema.CreatedAt,
		UpdatedAt: schema.UpdatedAt,
	}

	return &domain
}

//go:generate go run go.uber.org/mock/mockgen -source=servers.go -destination=mocks/mock_servers.go -package=mock_servers
type (
	Database struct {
		sqlite *sqliteclient.SqliteClient
	}

	Interface interface {
		GetServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error)
		CreateServerWithConfig(ctx context.Context, server domain.Server) (*domain.Server, error)
	}

	FxArgs struct {
		fx.In

		SqliteClient *sqliteclient.SqliteClient
	}
)

func New(args FxArgs) (*Database, error) {
	database := Database{
		sqlite: args.SqliteClient,
	}

	models := []any{
		(*SchemaServer)(nil),
		(*SchemaConfig)(nil),
	}

	for _, model := range models {
		_, err := args.SqliteClient.DB.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(context.Background())
		if err != nil {
			return nil, repoerrors.TranslateSqliteError(err)
		}
	}

	return &database, nil
}

func (database *Database) CreateServerWithConfig(
	ctx context.Context,
	server domain.Server,
) (*domain.Server, error) {
	if server.Config == nil {
		return nil, errors.New("server config is required")
	}
	if server.GuildID != server.Config.GuildID {
		return nil, repoerrors.NewErrorSqlInvalidData(errors.New("invalid guildID"))
	}

	serverSchema := NewSchemaServer(&server)
	configSchema := NewSchemaConfig(server.Config)

	err := database.sqlite.DB.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().
			Model(serverSchema).
			On("CONFLICT (guild_id) DO NOTHING").
			Returning("*").
			Exec(ctx, serverSchema); err != nil {
			return err
		}

		configSchema := NewSchemaConfig(server.Config)
		if _, err := tx.NewInsert().
			Model(configSchema).
			On("CONFLICT (guild_id) DO UPDATE").
			Set("prefix = EXCLUDED.prefix").
			Returning("*").
			Exec(ctx, configSchema); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	serverSchema.Config = configSchema
	return serverSchema.NewDomainServer(), nil
}

func (database *Database) GetServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error) {
	schema := new(SchemaServer)

	err := database.sqlite.DB.NewSelect().
		Model(schema).
		Relation("Config").
		Where("server.guild_id = ?", guildID).
		Scan(ctx)
	if err != nil {
		return nil, repoerrors.TranslateSqliteError(err)
	}

	if schema.Config == nil {
		return nil, repoerrors.TranslateSqliteError(sql.ErrNoRows)
	}

	return schema.NewDomainServer(), nil
}
