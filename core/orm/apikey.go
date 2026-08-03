package orm

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// SysAPIKey stores hashed API keys for external RPC auth (model: core.user.apikey).
type SysAPIKey struct {
	ID        int    `orm:"id"`
	UserID    int    `orm:"user_id"`
	Name      string `orm:"name"`
	KeyPrefix string `orm:"key_prefix"`
	KeyHash   string `orm:"key_hash"`
	Active    bool   `orm:"active"`
	CreateDate string `orm:"create_date"`
}

func (SysAPIKey) ModelName() string { return "core.user.apikey" }
func (SysAPIKey) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "user_id", Type: Many2One, Relation: "core.user", Required: true, Index: true, String: "User"},
		{Name: "name", Type: Char, Required: true, String: "Name"},
		{Name: "key_prefix", Type: Char, Required: true, String: "Prefix"},
		{Name: "key_hash", Type: Char, Required: true, String: "Hash"},
		{Name: "active", Type: Boolean, DefaultVal: true, String: "Active"},
		{Name: "create_date", Type: DateTime, Required: true, String: "Created"},
	}
}

// CoreUserLog records successful (and optionally failed) logins.
type CoreUserLog struct {
	ID         int    `orm:"id"`
	UserID     int    `orm:"user_id"`
	CreateDate string `orm:"create_date"`
	IP         string `orm:"ip"`
	Result     string `orm:"result"` // success | failure
}

func (CoreUserLog) ModelName() string { return "core.user.log" }
func (CoreUserLog) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "user_id", Type: Many2One, Relation: "core.user", Index: true, String: "User"},
		{Name: "create_date", Type: DateTime, Required: true, String: "When"},
		{Name: "ip", Type: Char, String: "IP"},
		{Name: "result", Type: Char, Required: true, DefaultVal: "success", String: "Result"},
	}
}

func init() {
	const kernel = "base"
	RegisterModelWithModule(SysAPIKey{}, kernel)
	RegisterModelWithModule(CoreUserLog{}, kernel)
}

// HashAPIKey returns a hex SHA-256 digest of the raw key.
func HashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// GenerateAPIKey creates a random key sk_<hex>. Returns raw key (show once), prefix, and hash.
func GenerateAPIKey() (raw, prefix, hash string, err error) {
	buf := make([]byte, 24)
	if _, err = rand.Read(buf); err != nil {
		return "", "", "", err
	}
	raw = "sk_" + hex.EncodeToString(buf)
	prefix = raw
	if len(prefix) > 11 {
		prefix = prefix[:11]
	}
	return raw, prefix, HashAPIKey(raw), nil
}

// UIDFromAPIKey resolves an active API key to a user id. Returns 0 if invalid.
func UIDFromAPIKey(ctx context.Context, raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" || DB == nil {
		return 0
	}
	hash := HashAPIKey(raw)
	tbl := GetTableName("core.user.apikey")
	var uid int
	var active bool
	err := DB.QueryRowContext(ctx,
		`SELECT user_id, active FROM `+tbl+` WHERE key_hash = $1 LIMIT 1`, hash,
	).Scan(&uid, &active)
	if err == sql.ErrNoRows || err != nil || !active || uid <= 0 {
		return 0
	}
	// Ensure user is active.
	var userActive bool
	err = DB.QueryRowContext(ctx,
		`SELECT active FROM `+GetTableName("core.user")+` WHERE id = $1`, uid,
	).Scan(&userActive)
	if err != nil || !userActive {
		return 0
	}
	return uid
}

// CreateAPIKeyForUser inserts a new key row. Returns the raw key (caller must show once).
func CreateAPIKeyForUser(ctx context.Context, userID int, name string) (raw string, err error) {
	if userID <= 0 {
		return "", fmt.Errorf("user id required")
	}
	raw, prefix, hash, err := GenerateAPIKey()
	if err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "API Key"
	}
	_, err = Create(ctx, SysAPIKey{}, map[string]interface{}{
		"user_id":     userID,
		"name":        name,
		"key_prefix":  prefix,
		"key_hash":    hash,
		"active":      true,
		"create_date": time.Now().UTC().Format(time.RFC3339),
	})
	return raw, err
}

// AppendUserLog writes a login audit row (best-effort).
func AppendUserLog(ctx context.Context, userID int, ip, result string) {
	if DB == nil {
		return
	}
	result = strings.TrimSpace(result)
	if result == "" {
		result = "success"
	}
	vals := map[string]interface{}{
		"create_date": time.Now().UTC().Format(time.RFC3339),
		"ip":          strings.TrimSpace(ip),
		"result":      result,
	}
	if userID > 0 {
		vals["user_id"] = userID
	}
	bypass := ContextWithBypass(ctx, true)
	_, _ = Create(bypass, CoreUserLog{}, vals)
}
