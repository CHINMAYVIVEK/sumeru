package parser

import "strings"

// listOpenAttrDisablesRowNavigation is true when the arch attribute open="..." disables
// opening the form from a list row (open="false", "0", "off", "no").
func listOpenAttrDisablesRowNavigation(open string) bool {
	s := strings.TrimSpace(strings.ToLower(open))
	switch s {
	case "0", "false", "off", "no":
		return true
	default:
		return false
	}
}

func applyListOpenFlag(v *View) {
	if v == nil {
		return
	}
	if strings.ToLower(strings.TrimSpace(v.Type)) == "list" {
		v.ListNoRowOpen = listOpenAttrDisablesRowNavigation(v.ListOpenAttr)
	}
}
