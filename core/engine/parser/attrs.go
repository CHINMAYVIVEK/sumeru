package parser

import "strings"

// IsTruthyAttr reports whether an XML attribute is a recognized true value.
func IsTruthyAttr(raw string) bool {
	s := strings.ToLower(strings.TrimSpace(raw))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}
