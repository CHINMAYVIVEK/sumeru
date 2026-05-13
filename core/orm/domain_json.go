package orm

import (
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

// SubstituteDomainUID replaces string "$uid" in domain values with uid.
func SubstituteDomainUID(domain [][]interface{}, uid int) [][]interface{} {
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
		if s, ok := v.(string); ok && s == "$uid" {
			out[i] = []interface{}{d[0], d[1], int64(uid)}
			continue
		}
		out[i] = d
	}
	return out
}

// RecordMatchesDomain evaluates AND of triples for =, !=, in (value list).
func RecordMatchesDomain(rec map[string]interface{}, domain [][]interface{}) bool {
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
		return cellTruthy(dbVal) == w
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

func cellTruthy(v interface{}) bool {
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

// MergeDomains concatenates domains with AND semantics for SQL builder.
func MergeDomains(parts ...[][]interface{}) [][]interface{} {
	var out [][]interface{}
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}
