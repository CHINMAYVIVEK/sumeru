package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ParseDomainJSON parses a JSON array of [field, operator, value] triples into [][]interface{}.
// Supported operators: =, !=, in. Value may be number, string, bool, or JSON array for "in".
func ParseDomainJSON(raw string) ([][]interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil, nil
	}
	var triples [][]interface{}
	if err := json.Unmarshal([]byte(raw), &triples); err != nil {
		return nil, fmt.Errorf("domain JSON: %w", err)
	}
	out := make([][]interface{}, 0, len(triples))
	for _, t := range triples {
		if len(t) != 3 {
			return nil, fmt.Errorf("domain triple must have 3 elements: %v", t)
		}
		field, ok := t[0].(string)
		if !ok {
			return nil, fmt.Errorf("domain field must be string: %v", t[0])
		}
		op, ok := t[1].(string)
		if !ok {
			return nil, fmt.Errorf("domain op must be string: %v", t[1])
		}
		out = append(out, []interface{}{field, op, t[2]})
	}
	return out, nil
}

// ResolveDomainXMLRefs replaces many2one-style string values that look like XML ids
// (module.name) with their resolved integer ids.
func ResolveDomainXMLRefs(ctx context.Context, domain [][]interface{}) [][]interface{} {
	if len(domain) == 0 {
		return domain
	}
	out := make([][]interface{}, len(domain))
	for i, d := range domain {
		if len(d) != 3 {
			out[i] = d
			continue
		}
		s, ok := d[2].(string)
		if !ok {
			out[i] = d
			continue
		}
		s = strings.TrimSpace(s)
		if s == "" || !strings.Contains(s, ".") || strings.HasPrefix(s, "$") {
			out[i] = d
			continue
		}
		id, _, err := ResolveXmlId(ctx, s)
		if err != nil || id <= 0 {
			out[i] = d
			continue
		}
		out[i] = []interface{}{d[0], d[1], id}
	}
	return out
}

// SubstituteDomainUID replaces string "$uid" in domain values with uid.
// Prefer SubstituteDomainContext for ABAC tokens ($company_id, $company_ids).
func SubstituteDomainUID(domain [][]interface{}, uid int) [][]interface{} {
	return SubstituteDomainContext(domain, DomainContext{UID: uid})
}

// DomainContext holds ABAC substitution tokens for record-rule domains.
type DomainContext struct {
	UID        int
	CompanyID  int64
	CompanyIDs []int64
}

// SubstituteDomainContext replaces $uid, $company_id, and $company_ids in domain values.
func SubstituteDomainContext(domain [][]interface{}, dc DomainContext) [][]interface{} {
	if len(domain) == 0 {
		return domain
	}
	out := make([][]interface{}, len(domain))
	for i, d := range domain {
		if len(d) != 3 {
			out[i] = d
			continue
		}
		v := d[2]
		s, ok := v.(string)
		if !ok {
			out[i] = d
			continue
		}
		switch s {
		case "$uid", "uid":
			out[i] = []interface{}{d[0], d[1], int64(dc.UID)}
		case "$company_id", "company_id":
			out[i] = []interface{}{d[0], d[1], dc.CompanyID}
		case "$company_ids", "company_ids":
			arr := make([]interface{}, len(dc.CompanyIDs))
			for j, id := range dc.CompanyIDs {
				arr[j] = id
			}
			out[i] = []interface{}{d[0], d[1], arr}
		default:
			out[i] = d
		}
	}
	return out
}

// RecordMatchesDomain evaluates AND of triples for =, !=, in (value list).
// Prefix Polish "|" markers OR the following leaf triples (same shape as ApplicableRuleDomains).
func RecordMatchesDomain(rec map[string]interface{}, domain [][]interface{}) bool {
	orLeaves := 0
	for _, d := range domain {
		if len(d) == 1 && fmt.Sprint(d[0]) == "|" {
			orLeaves++
			continue
		}
		break
	}
	if orLeaves > 0 {
		leaves := domain[orLeaves:]
		if len(leaves) != orLeaves+1 {
			return false
		}
		for _, leaf := range leaves {
			if RecordMatchesDomain(rec, [][]interface{}{leaf}) {
				return true
			}
		}
		return false
	}
	for _, d := range domain {
		if len(d) != 3 {
			return false
		}
		field, _ := d[0].(string)
		op, _ := d[1].(string)
		cell, ok := rec[field]
		if !ok {
			cell = nil
		}
		switch op {
		case "=":
			if !valuesEqual(cell, d[2]) {
				return false
			}
		case "!=":
			if valuesEqual(cell, d[2]) {
				return false
			}
		case "in":
			arr, ok := d[2].([]interface{})
			if !ok {
				return false
			}
			found := false
			for _, x := range arr {
				if valuesEqual(cell, x) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func valuesEqual(dbVal interface{}, want interface{}) bool {
	if want == nil && (dbVal == nil || AsString(dbVal) == "") {
		return true
	}
	switch w := want.(type) {
	case float64:
		cv, ok := CoerceInt64(dbVal)
		if !ok {
			f2, ok2 := toFloat64(dbVal)
			return ok2 && f2 == w
		}
		return float64(cv) == w
	case bool:
		return AsBool(dbVal) == w
	case string:
		return strings.TrimSpace(AsString(dbVal)) == strings.TrimSpace(w)
	default:
		cv, ok := CoerceInt64(dbVal)
		if ok {
			wi, ok2 := CoerceInt64(want)
			return ok2 && cv == wi
		}
		return AsString(dbVal) == AsString(want)
	}
}

// AsBool coerces a database cell value to bool.
func AsBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	case int32:
		return t != 0
	case int:
		return t != 0
	case float64:
		return t != 0
	case []byte:
		s := strings.ToLower(strings.TrimSpace(string(t)))
		return s == "t" || s == "true" || s == "1"
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		return s == "t" || s == "true" || s == "1"
	default:
		return false
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	default:
		s := strings.TrimSpace(AsString(v))
		if s == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(s, 64)
		return f, err == nil
	}
}

