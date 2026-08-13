package orm

import (
	"fmt"
	"strconv"
	"strings"
)

// WriteOp selects PrepareValues behavior for create vs write.
type WriteOp string

const (
	WriteOpCreate WriteOp = "create"
	WriteOpWrite  WriteOp = "write"
)

// PrepareOptions controls unknown-key handling.
type PrepareOptions struct {
	// StrictUnknown rejects undeclared field keys. When false, unknown keys are dropped silently
	// (historical Update behavior). Create uses StrictUnknown=true.
	StrictUnknown bool
}

// PrepareValues whitelists model fields, coerces types, and validates required fields on create.
// Relational virtual fields (Many2Many, One2Many) are excluded from SQL column maps.
func PrepareValues(model Model, values map[string]interface{}, op WriteOp, opts PrepareOptions) (map[string]interface{}, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}
	if values == nil {
		values = map[string]interface{}{}
	}
	defs := map[string]FieldDefinition{}
	for _, f := range model.Fields() {
		if f.Name == "" || f.Name == "id" {
			continue
		}
		defs[f.Name] = f
	}

	out := make(map[string]interface{}, len(values))
	for k, v := range values {
		if k == "id" {
			continue
		}
		fd, ok := defs[k]
		if !ok {
			if opts.StrictUnknown {
				return nil, fmt.Errorf("unknown field %q on model %s", k, model.ModelName())
			}
			continue
		}
		switch fd.Type {
		case Many2Many, One2Many:
			// Stored via relation tables; not INSERT/UPDATE columns.
			continue
		}
		cv, err := coerceFieldValue(fd, v)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", k, err)
		}
		out[k] = cv
	}

	if op == WriteOpCreate {
		for name, fd := range defs {
			if fd.Type == Many2Many || fd.Type == One2Many {
				continue
			}
			if !fd.Required {
				continue
			}
			if _, ok := out[name]; ok {
				continue
			}
			if fd.DefaultVal != nil {
				cv, err := coerceFieldValue(fd, fd.DefaultVal)
				if err != nil {
					return nil, fmt.Errorf("field %s default: %w", name, err)
				}
				out[name] = cv
				continue
			}
			return nil, fmt.Errorf("required field %q missing on model %s", name, model.ModelName())
		}
	}
	return out, nil
}

func coerceFieldValue(fd FieldDefinition, v interface{}) (interface{}, error) {
	if v == nil {
		if fd.Required {
			return nil, fmt.Errorf("required value is null")
		}
		return nil, nil
	}
	switch fd.Type {
	case Boolean:
		switch t := v.(type) {
		case bool:
			return t, nil
		case string:
			s := strings.ToLower(strings.TrimSpace(t))
			if s == "true" || s == "1" || s == "yes" {
				return true, nil
			}
			if s == "false" || s == "0" || s == "no" || s == "" {
				return false, nil
			}
			return nil, fmt.Errorf("invalid boolean %q", t)
		case int, int32, int64, float32, float64:
			n, _ := CoerceInt64(v)
			return n != 0, nil
		default:
			return AsBool(v), nil
		}
	case Integer, Many2One:
		n, ok := CoerceInt64(v)
		if !ok {
			s := strings.TrimSpace(AsString(v))
			if s == "" {
				if fd.Required {
					return nil, fmt.Errorf("required integer empty")
				}
				return nil, nil
			}
			return nil, fmt.Errorf("invalid integer %v", v)
		}
		return int(n), nil
	case Float, Numeric:
		switch t := v.(type) {
		case float64:
			return t, nil
		case float32:
			return float64(t), nil
		case int, int32, int64:
			n, _ := CoerceInt64(v)
			return float64(n), nil
		case string:
			s := strings.TrimSpace(t)
			if s == "" {
				return nil, nil
			}
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid float %q", t)
			}
			return f, nil
		default:
			s := strings.TrimSpace(AsString(v))
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid float %v", v)
			}
			return f, nil
		}
	case Selection:
		s := strings.TrimSpace(AsString(v))
		if s == "" {
			if fd.Required {
				return nil, fmt.Errorf("required selection empty")
			}
			return "", nil
		}
		if len(fd.Selection) == 0 {
			return s, nil
		}
		for _, opt := range fd.Selection {
			if len(opt) >= 1 && opt[0] == s {
				return s, nil
			}
		}
		return nil, fmt.Errorf("invalid selection %q", s)
	case Char, Text, Date, DateTime, Json:
		return AsString(v), nil
	default:
		return v, nil
	}
}
