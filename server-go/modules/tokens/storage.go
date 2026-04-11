package tokens

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/raghavendra-talur/go-vite-template/server-go/db"
)

func mustParseTime(s string) time.Time {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05Z"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parseNullableTime(ns sql.NullString) *time.Time {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	t := mustParseTime(ns.String)
	if t.IsZero() {
		return nil
	}
	return &t
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "rb_" + hex.EncodeToString(b)
}

const tokenColumns = "id, name, token, created_at, last_used_at"

func scanToken(row interface{ Scan(dest ...any) error }) (*Token, error) {
	var t Token
	var createdAt string
	var lastUsedAt sql.NullString

	err := row.Scan(&t.ID, &t.Name, &t.Token, &createdAt, &lastUsedAt)
	if err != nil {
		return nil, err
	}

	t.CreatedAt = mustParseTime(createdAt)
	t.LastUsedAt = parseNullableTime(lastUsedAt)
	return &t, nil
}

func Create(ctx context.Context, name string) (*Token, error) {
	token := generateToken()
	row := db.DB.QueryRowContext(ctx,
		"INSERT INTO api_tokens (name, token) VALUES (?, ?) RETURNING "+tokenColumns,
		name, token)
	return scanToken(row)
}

func GetAll(ctx context.Context) ([]Token, error) {
	rows, err := db.DB.QueryContext(ctx,
		"SELECT id, name, '' as token, created_at, last_used_at FROM api_tokens ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := []Token{}
	for rows.Next() {
		t, err := scanToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, *t)
	}
	return tokens, rows.Err()
}

func ValidateToken(ctx context.Context, token string) bool {
	var id string
	err := db.DB.QueryRowContext(ctx,
		"SELECT id FROM api_tokens WHERE token = ?", token).Scan(&id)
	if err != nil {
		return false
	}
	// Update last_used_at in background
	db.DB.ExecContext(ctx,
		"UPDATE api_tokens SET last_used_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = ?", id)
	return true
}

func Delete(ctx context.Context, id string) (bool, error) {
	result, err := db.DB.ExecContext(ctx, "DELETE FROM api_tokens WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	n, _ := result.RowsAffected()
	return n > 0, nil
}
