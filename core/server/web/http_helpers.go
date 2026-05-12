package web

import "fmt"

func intFromDB(v interface{}) (int, bool) {
	switch t := v.(type) {
	case int64:
		return int(t), true
	case int32:
		return int(t), true
	case int:
		return t, true
	case float64:
		return int(t), true
	case []byte:
		var n int
		_, err := fmt.Sscanf(string(t), "%d", &n)
		return n, err == nil
	default:
		return 0, false
	}
}

func splitViewModes(viewMode string) []string {
	var out []string
	for _, p := range splitComma(viewMode) {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func splitComma(s string) []string {
	var parts []string
	for _, raw := range splitByChar(s, ',') {
		if t := trimSpace(raw); t != "" {
			parts = append(parts, t)
		}
	}
	return parts
}

func splitByChar(s string, sep rune) []string {
	var cur []rune
	var out []string
	for _, r := range s {
		if r == sep {
			out = append(out, string(cur))
			cur = nil
			continue
		}
		cur = append(cur, r)
	}
	out = append(out, string(cur))
	return out
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func normalizeViewMode(mode string) string {
	m := trimSpace(mode)
	if m == "" {
		return m
	}
	switch m {
	case "list":
		return "tree"
	default:
		return m
	}
}
