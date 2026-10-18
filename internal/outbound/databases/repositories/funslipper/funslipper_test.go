package funslipper

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/infra/sqliteclient"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func newTestDB(t *testing.T) *Database {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqldb.Close() })
	sqldb.SetMaxOpenConns(1)

	db, err := New(FxArgs{SqliteClient: &sqliteclient.SqliteClient{DB: bun.NewDB(sqldb, sqlitedialect.New())}})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestIncrementCount(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	first, err := db.IncrementCount(ctx, domain.FunSlipper{GuildID: "g1", UserID: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if first.Count != 1 {
		t.Fatalf("first increment should be 1, got %d", first.Count)
	}

	second, err := db.IncrementCount(ctx, domain.FunSlipper{GuildID: "g1", UserID: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if second.Count != 2 {
		t.Fatalf("second increment should be 2, got %d", second.Count)
	}

	otherGuild, err := db.IncrementCount(ctx, domain.FunSlipper{GuildID: "g2", UserID: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if otherGuild.Count != 1 {
		t.Fatalf("same user in another guild should start at 1, got %d", otherGuild.Count)
	}
}
