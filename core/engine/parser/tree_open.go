package parser

import "strings"

// treeOpenAttrDisablesRowNavigation is true when the arch attribute open="..." disables
// opening the form from a list row (same convention as Odoo: open="false", "0", "off", "no").
func treeOpenAttrDisablesRowNavigation(open string) bool {
	s := strings.TrimSpace(strings.ToLower(open))
	switch s {
	case "0", "false", "off", "no":
		return true
	default:
		return false
	}
}

func applyTreeListOpenFlag(v *View) {
	if v == nil {
		return
	}
	t := strings.ToLower(strings.TrimSpace(v.Type))
	if t == "tree" || t == "list" {
		v.TreeNoRowOpen = treeOpenAttrDisablesRowNavigation(v.TreeOpenAttr)
	}
}
