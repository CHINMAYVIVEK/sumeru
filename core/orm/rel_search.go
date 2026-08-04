package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// RelNameSearch returns up to limit id/name pairs for a comodel (name or login ILIKE).
func RelNameSearch(ctx context.Context, modelName, query string, limit int) ([]map[string]interface{}, error) {
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return nil, fmt.Errorf("model required")
	}
	m, ok := Registry[modelName]
	if !ok || m == nil {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "read"); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	nameCol := "name"
	hasName, hasLogin := false, false
	for _, f := range m.Fields() {
		switch f.Name {
		case "name":
			hasName = true
		case "login":
			hasLogin = true
		}
	}
	if !hasName && hasLogin {
		nameCol = "login"
	} else if !hasName && !hasLogin {
		return nil, fmt.Errorf("model %s has no name/login field", modelName)
	}
	tbl := GetTableName(modelName)
	q := strings.TrimSpace(query)
	var rows *sql.Rows
	var err error
	if q == "" {
		sqlQ := fmt.Sprintf(`SELECT id, COALESCE(NULLIF(TRIM(%s::text), ''), '') FROM %s ORDER BY id DESC LIMIT $1`, quoteIdent(nameCol), tbl)
		rows, err = DB.QueryContext(ctx, sqlQ, limit)
	} else {
		pat := "%" + q + "%"
		if hasName && hasLogin {
			sqlQ := fmt.Sprintf(`SELECT id, COALESCE(NULLIF(TRIM(name), ''), NULLIF(TRIM(login), ''), '') FROM %s WHERE name ILIKE $1 OR login ILIKE $1 ORDER BY id DESC LIMIT $2`, tbl)
			rows, err = DB.QueryContext(ctx, sqlQ, pat, limit)
		} else {
			sqlQ := fmt.Sprintf(`SELECT id, COALESCE(NULLIF(TRIM(%s::text), ''), '') FROM %s WHERE %s::text ILIKE $1 ORDER BY id DESC LIMIT $2`, quoteIdent(nameCol), tbl, quoteIdent(nameCol))
			rows, err = DB.QueryContext(ctx, sqlQ, pat, limit)
		}
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{"id": id, "name": name})
	}
	return out, rows.Err()
}

// ResolveInverseOne2ManyField finds the Many2One on comodel pointing at parentModel.
func ResolveInverseOne2ManyField(parentModel, comodel string) string {
	m, ok := Registry[comodel]
	if !ok || m == nil {
		return ""
	}
	var fallback string
	for _, f := range m.Fields() {
		if f.Type != Many2One || f.Relation != parentModel {
			continue
		}
		if strings.HasSuffix(f.Name, "_id") {
			return f.Name
		}
		if fallback == "" {
			fallback = f.Name
		}
	}
	return fallback
}

// DisplayNameForID returns a short label for a related record id.
func DisplayNameForID(ctx context.Context, modelName string, id int) string {
	if id <= 0 || strings.TrimSpace(modelName) == "" {
		return ""
	}
	rec, err := SearchOne(ctx, modelName, map[string]interface{}{"id": id})
	if err != nil {
		return fmt.Sprintf("%d", id)
	}
	if n := strings.TrimSpace(AsString(rec["name"])); n != "" {
		return n
	}
	if n := strings.TrimSpace(AsString(rec["login"])); n != "" {
		return n
	}
	return fmt.Sprintf("%d", id)
}
