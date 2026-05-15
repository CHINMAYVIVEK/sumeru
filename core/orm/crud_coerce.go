package orm

import (
	"fmt"
	"strings"
)

// CoerceInt64 reads numeric values from database drivers into int64.
func CoerceInt64(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case int32:
		return int64(t), true
	case int:
		return int64(t), true
	case uint64:
		return int64(t), true
	case uint32:
		return int64(t), true
	case float64:
		return int64(t), true
	case float32:
		return int64(t), true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, false
		}
		var n int64
		_, err := fmt.Sscanf(s, "%d", &n)
		return n, err == nil
	case []byte:
		var n int64
		_, err := fmt.Sscanf(string(t), "%d", &n)
		return n, err == nil
	default:
		return 0, false
	}
}

// AsString coerces common database driver values to string.
func AsString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case nil:
		return ""
	default:
		return fmt.Sprint(t)
	}
}
