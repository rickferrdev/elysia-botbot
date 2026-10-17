package repoerrors_test

import (
	"database/sql"
	"errors"
	"testing"
	"unsafe"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
	"modernc.org/sqlite"
)

type fakeSqliteError struct {
	msg  string
	code int
}

func newSqliteError(code int) error {
	fake := fakeSqliteError{msg: "sqlite error", code: code}
	return (*sqlite.Error)(unsafe.Pointer(&fake))
}

func TestTranslateSqliteError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if err := repoerrors.TranslateSqliteError(nil); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("sql.ErrNoRows", func(t *testing.T) {
		err := repoerrors.TranslateSqliteError(sql.ErrNoRows)
		if !core.IsCode(err, repoerrors.CodeSqlNoRows) {
			t.Errorf("expected CodeSqlNoRows, got %v", err)
		}
	})

	t.Run("unknown non-sqlite error", func(t *testing.T) {
		err := repoerrors.TranslateSqliteError(errors.New("some unexpected error"))
		if !core.IsCode(err, repoerrors.CodeSqlUnknown) {
			t.Errorf("expected CodeSqlUnknown, got %v", err)
		}
	})

	t.Run("sqlite mapped codes", func(t *testing.T) {
		tests := []struct {
			name         string
			sqliteCode   int
			expectedCode core.ErrorCode
		}{
			{"conflict", 2067, repoerrors.CodeSqlConflict},
			{"foreign_key", 787, repoerrors.CodeSqlForeignKey},
			{"required", 1299, repoerrors.CodeSqlRequired},
			{"busy", 5, repoerrors.CodeSqlBusy},
			{"constraint_masked", 19, repoerrors.CodeSqlConstraint},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				fakeErr := newSqliteError(tc.sqliteCode)
				translated := repoerrors.TranslateSqliteError(fakeErr)
				if !core.IsCode(translated, tc.expectedCode) {
					t.Errorf("code %d: expected %s, got %v", tc.sqliteCode, tc.expectedCode, translated)
				}
			})
		}
	})
}
