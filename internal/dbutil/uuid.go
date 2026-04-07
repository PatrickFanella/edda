// Package dbutil provides shared helpers for converting between domain types
// and database-specific types (pgtype, pgvector).
//
// Deprecated: New code should use github.com/PatrickFanella/game-master/internal/db
// directly. These functions are thin wrappers for backward compatibility.
package dbutil

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/db"
)

// FromPgtype converts a pgtype.UUID to a uuid.UUID.
// Invalid UUIDs are returned as uuid.Nil.
//
// Deprecated: Use db.FromPgUUID instead.
func FromPgtype(p pgtype.UUID) uuid.UUID {
	return db.FromPgUUID(p)
}

// ToPgtype converts a uuid.UUID to a pgtype.UUID.
// uuid.Nil is stored with Valid set to false.
//
// Deprecated: Use db.ToPgUUID instead.
func ToPgtype(u uuid.UUID) pgtype.UUID {
	return db.ToPgUUID(u)
}

// PgUUIDsToStrings converts a slice of pgtype.UUID to string representations,
// skipping any invalid entries.
//
// Deprecated: Use db.PgUUIDsToStrings instead.
func PgUUIDsToStrings(ids []pgtype.UUID) []string {
	return db.PgUUIDsToStrings(ids)
}

// UUIDsToPgtype converts a slice of uuid.UUID to pgtype.UUID.
//
// Deprecated: Use db.UUIDsToPgtype instead.
func UUIDsToPgtype(ids []uuid.UUID) []pgtype.UUID {
	return db.UUIDsToPgtype(ids)
}
