package orm

import (
	"context"
	"fmt"
	"strings"
)

// RelNameSearchFiltered returns name_search rows with an optional equality filter (filterField = filterID).
func RelNameSearchFiltered(ctx context.Context, modelName, query string, limit int, filterField string, filterID int64) ([]map[string]interface{}, error) {
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
	if limit <= 0 {
		limit = 20
	}
	if limit > 500 {
		limit = 500
	}
	nameCol := "name"
	hasName, hasLogin, hasPhoneCode := false, false, false
	validFilter := false
	for _, f := range m.Fields() {
		switch f.Name {
		case "name":
			hasName = true
		case "login":
			hasLogin = true
		case "phone_code":
			hasPhoneCode = true
		}
		if filterField != "" && f.Name == filterField {
			validFilter = true
		}
	}
	if filterField != "" && !validFilter {
		return nil, fmt.Errorf("invalid filter field %q on %s", filterField, modelName)
	}
	if !hasName && hasLogin {
		nameCol = "login"
	} else if !hasName && !hasLogin {
		return nil, fmt.Errorf("model %s has no name/login field", modelName)
	}
	if filterField != "" && filterID <= 0 {
		return []map[string]interface{}{}, nil
	}

	var baseDomain [][]interface{}
	if filterField != "" && filterID > 0 {
		baseDomain = append(baseDomain, []interface{}{filterField, "=", filterID})
	}
	ruleWhere, ruleArgs, err := BuildWhereWithRecordRules(ctx, uid, modelName, "read", baseDomain)
	if err != nil {
		return nil, err
	}

	tbl, err := QuotedTableForModel(modelName)
	if err != nil {
		return nil, err
	}
	nameIdent, err := QuotedColumnForModel(modelName, nameCol)
	if err != nil {
		return nil, err
	}
	selectCols := fmt.Sprintf(`id, COALESCE(NULLIF(TRIM(%s::text), ''), '')`, nameIdent)
	if hasPhoneCode {
		phoneIdent, err := QuotedColumnForModel(modelName, "phone_code")
		if err != nil {
			return nil, err
		}
		selectCols += fmt.Sprintf(`, COALESCE(NULLIF(TRIM(%s::text), ''), '')`, phoneIdent)
	}
	q := strings.TrimSpace(query)
	customWhere := []string{}
	customArgs := []interface{}{}
	n := len(ruleArgs) + 1
	if q != "" {
		pat := "%" + q + "%"
		if hasName && hasLogin {
			loginIdent, err := QuotedColumnForModel(modelName, "login")
			if err != nil {
				return nil, err
			}
			nameForLike, err := QuotedColumnForModel(modelName, "name")
			if err != nil {
				return nil, err
			}
			customWhere = append(customWhere, fmt.Sprintf(`(%s ILIKE $%d OR %s ILIKE $%d)`, nameForLike, n, loginIdent, n))
		} else {
			customWhere = append(customWhere, fmt.Sprintf(`%s::text ILIKE $%d`, nameIdent, n))
		}
		customArgs = append(customArgs, pat)
		n++
	}
	whereSQL := ruleWhere
	args := append([]interface{}(nil), ruleArgs...)
	if len(customWhere) > 0 {
		customSQL := strings.Join(customWhere, " AND ")
		shifted, _ := shiftPlaceholders(customSQL, len(args)+1)
		if whereSQL == "1=1" && len(baseDomain) == 0 {
			whereSQL = shifted
		} else {
			whereSQL = whereSQL + " AND (" + shifted + ")"
		}
		args = append(args, customArgs...)
	}
	sqlQ := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, selectCols, tbl, whereSQL)
	sqlQ += fmt.Sprintf(` ORDER BY %s ASC, id ASC LIMIT $%d`, nameIdent, len(args)+1)
	args = append(args, limit)

	rows, err := DB.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		item := map[string]interface{}{}
		if hasPhoneCode {
			var phoneCode string
			if err := rows.Scan(&id, &name, &phoneCode); err != nil {
				return nil, err
			}
			item["phone_code"] = phoneCode
		} else {
			if err := rows.Scan(&id, &name); err != nil {
				return nil, err
			}
		}
		item["id"] = id
		item["name"] = name
		out = append(out, item)
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
