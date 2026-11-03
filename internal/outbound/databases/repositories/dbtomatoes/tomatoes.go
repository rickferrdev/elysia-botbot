package dbtomatoes

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

//go:generate mockgen -source=tomatoes.go -destination=mocks/mock_tomatoes.go -package=mock_tomatoes
type (
	Database struct {
		sqlite *sqliteclient.SqliteClient
	}

	Interface interface {
		Increment(ctx context.Context, guildID, userID string, value int) (int, error)
		Decrement(ctx context.Context, guildID, userID string, value int) (int, error)
		GetInventory(ctx context.Context, guildID, userID string) (int, error)
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

	_, err := args.SqliteClient.DB.NewCreateTable().
		Model((*SchemaTomatoes)(nil)).
		IfNotExists().
		Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return &database, nil
}

func (database *Database) Increment(ctx context.Context, guildID, userID string, value int) (int, error) {
	var inventory int
	schema := SchemaTomatoes{
		UserID:    userID,
		GuildID:   guildID,
		Inventory: value,
	}

	err := database.sqlite.DB.NewInsert().
		Model(&schema).
		Where("guild_id = ? AND user_id = ?", guildID, userID).
		On("CONFLICT (user_id, guild_id) DO UPDATE").
		Set("inventory = inventory + ?", value).
		Returning("tomatoes.inventory").
		Scan(ctx, &inventory)
	if err != nil {
		return inventory, err
	}

	return inventory, nil
}

func (database *Database) Decrement(ctx context.Context, guildID, userID string, value int) (int, error) {
	var inventory int
	schema := SchemaTomatoes{
		UserID:    userID,
		GuildID:   guildID,
		Inventory: value,
	}

	err := database.sqlite.DB.NewInsert().
		Model(&schema).
		Where("guild_id = ? AND user_id = ?", guildID, userID).
		On("CONFLICT (user_id, guild_id) DO UPDATE").
		Set("inventory = inventory - ?", value).
		Returning("tomatoes.inventory").
		Scan(ctx, &inventory)
	if err != nil {
		return inventory, err
	}

	return inventory, nil
}

func (database *Database) GetInventory(ctx context.Context, guildID, userID string) (int, error) {
	var inventory int
	err := database.sqlite.DB.NewSelect().
		Model((*SchemaTomatoes)(nil)).
		Where("guild_id = ? AND user_id = ?", guildID, userID).
		Column("inventory").
		Scan(ctx, &inventory)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, repoerrors.NewErrorSqlNoRows(err)
		}
		return inventory, repoerrors.TranslateSqliteError(err)
	}

	return inventory, nil
}
