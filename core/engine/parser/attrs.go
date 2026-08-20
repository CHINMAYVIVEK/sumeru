package parser

import "strings"

// IsTruthyAttr reports whether an XML attribute is a recognized true value.
func IsTruthyAttr(raw string) bool {
	s := strings.ToLower(strings.TrimSpace(raw))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

// IsFalsyAttr reports whether an XML attribute is a recognized false value.
func IsFalsyAttr(raw string) bool {
	s := strings.ToLower(strings.TrimSpace(raw))
	return s == "0" || s == "false" || s == "no" || s == "off"
}

// AttrLiteralOrExpr splits a modifier attribute into a boolean literal or an expression.
// Empty and recognized true/false values are literals; anything else is an expression.
func AttrLiteralOrExpr(raw string) (literal bool, truthy bool, expr string) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return true, false, ""
	}
	if IsTruthyAttr(s) {
		return true, true, ""
	}
	if IsFalsyAttr(s) {
		return true, false, ""
	}
	return false, false, s
}
