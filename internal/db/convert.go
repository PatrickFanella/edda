package db

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ToPgUUID converts a uuid.UUID to a pgtype.UUID.
// uuid.Nil is stored with Valid set to false.
func ToPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: id != uuid.Nil}
}

// FromPgUUID converts a pgtype.UUID to a uuid.UUID.
// Invalid UUIDs are returned as uuid.Nil.
func FromPgUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return uuid.UUID(id.Bytes)
}

// ToPgText converts a string to a pgtype.Text.
// An empty string is stored with Valid set to false.
func ToPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// FromPgText converts a pgtype.Text to a string.
// Invalid text is returned as an empty string.
func FromPgText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// UUIDsToPgtype converts a slice of uuid.UUID to pgtype.UUID.
func UUIDsToPgtype(ids []uuid.UUID) []pgtype.UUID {
	if len(ids) == 0 {
		return nil
	}
	out := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		out[i] = ToPgUUID(id)
	}
	return out
}

// PgUUIDsToStrings converts a slice of pgtype.UUID to string representations,
// skipping any invalid entries.
func PgUUIDsToStrings(ids []pgtype.UUID) []string {
	if len(ids) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if !id.Valid {
			continue
		}
		out = append(out, FromPgUUID(id).String())
	}
	return out
}
