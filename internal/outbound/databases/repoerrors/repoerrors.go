package repoerrors

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"modernc.org/sqlite"
)

const (
	CodeSqlNoRows          core.ErrorCode = "SQL_NO_ROWS"
	CodeSqlUnknown         core.ErrorCode = "SQL_UNKNOWN"
	CodeSqlConflict        core.ErrorCode = "SQL_CONFLICT"
	CodeSqlForeignKey      core.ErrorCode = "SQL_FOREIGN_KEY"
	CodeSqlRequired        core.ErrorCode = "SQL_REQUIRED"
	CodeSqlInvalidData     core.ErrorCode = "SQL_INVALID_DATA"
	CodeSqlConstraint      core.ErrorCode = "SQL_CONSTRAINT"
	CodeSqlBusy            core.ErrorCode = "SQL_BUSY"
	CodeSqlLocked          core.ErrorCode = "SQL_LOCKED"
	CodeSqlReadOnly        core.ErrorCode = "SQL_READ_ONLY"
	CodeSqlUnavailable     core.ErrorCode = "SQL_UNAVAILABLE"
	CodeSqlCorrupt         core.ErrorCode = "SQL_CORRUPT"
	CodeSqlInvalidDatabase core.ErrorCode = "SQL_INVALID_DATABASE"
	CodeSqlStorageFull     core.ErrorCode = "SQL_STORAGE_FULL"
	CodeSqlIO              core.ErrorCode = "SQL_IO"
	CodeSqlSyntax          core.ErrorCode = "SQL_SYNTAX"
	CodeSqlInterrupted     core.ErrorCode = "SQL_INTERRUPTED"
	CodeSqlOutOfMemory     core.ErrorCode = "SQL_OUT_OF_MEMORY"
)

func NewErrorSqlNoRows(cause error) error {
	if errors.Is(cause, sql.ErrNoRows) {
		return core.NewAppError("resource does not exist", CodeSqlNoRows, cause)
	}
	return NewErrorSqlUnknown(cause)
}

func NewErrorSqlUnknown(cause error) error {
	return core.NewAppError("unknown error occurred", CodeSqlUnknown, cause)
}

func NewErrorSqlConflict(cause error) error {
	return core.NewAppError("conflict occurred", CodeSqlConflict, cause)
}

func NewErrorSqlForeignKey(cause error) error {
	return core.NewAppError("foreign key constraint failed", CodeSqlForeignKey, cause)
}

func NewErrorSqlRequired(cause error) error {
	return core.NewAppError("required field is missing", CodeSqlRequired, cause)
}

func NewErrorSqlInvalidData(cause error) error {
	return core.NewAppError("invalid data", CodeSqlInvalidData, cause)
}

func NewErrorSqlConstraint(cause error) error {
	return core.NewAppError("database constraint violated", CodeSqlConstraint, cause)
}

func NewErrorSqlBusy(cause error) error {
	return core.NewAppError("database is busy", CodeSqlBusy, cause)
}

func NewErrorSqlLocked(cause error) error {
	return core.NewAppError("database is locked", CodeSqlLocked, cause)
}

func NewErrorSqlReadOnly(cause error) error {
	return core.NewAppError("database is read-only", CodeSqlReadOnly, cause)
}

func NewErrorSqlUnavailable(cause error) error {
	return core.NewAppError("database is unavailable", CodeSqlUnavailable, cause)
}

func NewErrorSqlCorrupt(cause error) error {
	return core.NewAppError("database is corrupted", CodeSqlCorrupt, cause)
}

func NewErrorSqlInvalidDatabase(cause error) error {
	return core.NewAppError("invalid database file", CodeSqlInvalidDatabase, cause)
}

func NewErrorSqlStorageFull(cause error) error {
	return core.NewAppError("database storage is full", CodeSqlStorageFull, cause)
}

func NewErrorSqlIO(cause error) error {
	return core.NewAppError("database I/O error", CodeSqlIO, cause)
}

func NewErrorSqlSyntax(cause error) error {
	return core.NewAppError("invalid SQL statement", CodeSqlSyntax, cause)
}

func NewErrorSqlInterrupted(cause error) error {
	return core.NewAppError("database operation interrupted", CodeSqlInterrupted, cause)
}

func NewErrorSqlOutOfMemory(cause error) error {
	return core.NewAppError("insufficient memory", CodeSqlOutOfMemory, cause)
}

func TranslateSqliteError(cause error) error {
	if cause == nil {
		return nil
	}

	if errors.Is(cause, sql.ErrNoRows) {
		return NewErrorSqlNoRows(cause)
	}

	var sqliteError *sqlite.Error
	if !errors.As(cause, &sqliteError) {
		return NewErrorSqlUnknown(cause)
	}

	switch sqliteError.Code() {
	case 2067, 1555, 2579:
		return NewErrorSqlConflict(cause)
	case 787:
		return NewErrorSqlForeignKey(cause)
	case 1299:
		return NewErrorSqlRequired(cause)
	case 275:
		return NewErrorSqlInvalidData(cause)
	case 5, 261, 517, 773:
		return NewErrorSqlBusy(cause)
	case 6, 262:
		return NewErrorSqlLocked(cause)
	case 8:
		return NewErrorSqlReadOnly(cause)
	case 14:
		return NewErrorSqlUnavailable(cause)
	case 11:
		return NewErrorSqlCorrupt(cause)
	case 26:
		return NewErrorSqlInvalidDatabase(cause)
	case 13:
		return NewErrorSqlStorageFull(cause)
	case 10:
		return NewErrorSqlIO(cause)
	case 1:
		return NewErrorSqlSyntax(cause)
	case 9:
		return NewErrorSqlInterrupted(cause)
	case 7:
		return NewErrorSqlOutOfMemory(cause)
	}

	switch sqliteError.Code() & 0xff {
	case 19:
		return NewErrorSqlConstraint(cause)
	case 5:
		return NewErrorSqlBusy(cause)
	case 6:
		return NewErrorSqlLocked(cause)
	case 8:
		return NewErrorSqlReadOnly(cause)
	case 10:
		return NewErrorSqlIO(cause)
	case 11:
		return NewErrorSqlCorrupt(cause)
	case 14:
		return NewErrorSqlUnavailable(cause)
	case 13:
		return NewErrorSqlStorageFull(cause)
	case 26:
		return NewErrorSqlInvalidDatabase(cause)
	case 7:
		return NewErrorSqlOutOfMemory(cause)
	}

	return fmt.Errorf("%w: %v", NewErrorSqlUnknown(cause), cause)
}
