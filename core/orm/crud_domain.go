package orm

import (
	"fmt"
	"strings"
)

func buildSearchWhereClause(domain [][]interface{}) (string, []interface{}, error) {
	if len(domain) == 0 {
		return "1=1", nil, nil
	}
	var parts []string
	var args []interface{}
	n := 1
	for _, d := range domain {
		if len(d) != 3 {
			return "", nil, fmt.Errorf("invalid domain clause %v", d)
		}
		field, ok := d[0].(string)
		if !ok || strings.TrimSpace(field) == "" {
			return "", nil, fmt.Errorf("domain field name")
		}
		op := strings.TrimSpace(strings.ToLower(fmt.Sprint(d[1])))
		col := quoteIdent(field)
		switch op {
		case "=":
			parts = append(parts, fmt.Sprintf("%s = $%d", col, n))
			args = append(args, d[2])
			n++
		case "!=":
			parts = append(parts, fmt.Sprintf("(%s IS DISTINCT FROM $%d)", col, n))
			args = append(args, d[2])
			n++
		case "in":
			list, ok := d[2].([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("operator in requires array value")
			}
			if len(list) == 0 {
				parts = append(parts, "FALSE")
				continue
			}
			ph := make([]string, len(list))
			for i := range list {
				ph[i] = fmt.Sprintf("$%d", n)
				args = append(args, list[i])
				n++
			}
			parts = append(parts, fmt.Sprintf("%s IN (%s)", col, strings.Join(ph, ",")))
		default:
			return "", nil, fmt.Errorf("unsupported domain operator %q", op)
		}
	}
	return strings.Join(parts, " AND "), args, nil
}
