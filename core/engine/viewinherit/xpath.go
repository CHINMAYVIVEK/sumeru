package viewinherit

import (
	"fmt"
	"regexp"
	"strings"
)

// xpathOp is one xpath directive from an inherit arch fragment.
type xpathOp struct {
	Expr     string
	Position string
	Inner    string
}

var xpathBlockRe = regexp.MustCompile(`(?s)<xpath\s+expr="([^"]+)"\s+position="([^"]+)"\s*>(.*?)</xpath>`)
var xpathBlockReSingle = regexp.MustCompile(`(?s)<xpath\s+expr='([^']+)'\s+position='([^']+)'\s*>(.*?)</xpath>`)
var fieldNameFromExpr = regexp.MustCompile(`@name=['"]([^'"]+)['"]`)
var dataWrapperRe = regexp.MustCompile(`(?s)^\s*<data[^>]*>(.*)</data>\s*$`)

// ApplyInheritArch parses <xpath> blocks in inheritFragment and applies them to parentArch.
// Supported: position after|before|inside|replace on //field[@name='x'] targets for flattened <view>…</view> arches.
func ApplyInheritArch(parentArch, inheritFragment string) (string, error) {
	arch := parentArch
	frag := stripDataWrapper(inheritFragment)
	ops := parseXPaths(frag)
	if len(ops) == 0 && strings.TrimSpace(frag) != "" {
		return arch, fmt.Errorf("no <xpath> blocks found in inherit arch")
	}
	for _, op := range ops {
		var err error
		arch, err = applyOne(arch, op)
		if err != nil {
			return arch, err
		}
	}
	return arch, nil
}

func parseXPaths(s string) []xpathOp {
	var out []xpathOp
	for _, re := range []*regexp.Regexp{xpathBlockRe, xpathBlockReSingle} {
		for _, m := range re.FindAllStringSubmatch(s, -1) {
			out = append(out, xpathOp{Expr: m[1], Position: m[2], Inner: m[3]})
		}
	}
	return out
}

func fieldNameFromXPathExpr(expr string) (string, error) {
	m := fieldNameFromExpr.FindStringSubmatch(expr)
	if len(m) < 2 {
		return "", fmt.Errorf("unsupported xpath expr (need //field[@name='…']): %q", expr)
	}
	return m[1], nil
}

func applyOne(arch string, op xpathOp) (string, error) {
	fn, err := fieldNameFromXPathExpr(strings.TrimSpace(op.Expr))
	if err != nil {
		return arch, err
	}
	pos := strings.ToLower(strings.TrimSpace(op.Position))
	inner := strings.TrimSpace(op.Inner)

	switch pos {
	case "after":
		re := regexp.MustCompile(`(<field\s+[^>]*\bname="` + regexp.QuoteMeta(fn) + `"[^>]*/>)`)
		loc := re.FindStringIndex(arch)
		if loc == nil {
			return arch, fmt.Errorf("inherit xpath: field %q not found for position=after", fn)
		}
		insert := inner
		if !strings.HasPrefix(strings.TrimSpace(insert), "<") {
			insert = "<field name=\"" + insert + "\"/>"
		}
		return arch[:loc[1]] + insert + arch[loc[1]:], nil
	case "before":
		re := regexp.MustCompile(`(<field\s+[^>]*\bname="` + regexp.QuoteMeta(fn) + `"[^>]*/>)`)
		loc := re.FindStringIndex(arch)
		if loc == nil {
			return arch, fmt.Errorf("inherit xpath: field %q not found for position=before", fn)
		}
		return arch[:loc[0]] + inner + arch[loc[0]:], nil
	case "replace":
		re := regexp.MustCompile(`<field\s+[^>]*\bname="` + regexp.QuoteMeta(fn) + `"[^>]*/>`)
		if !re.MatchString(arch) {
			return arch, fmt.Errorf("inherit xpath: field %q not found for position=replace", fn)
		}
		return re.ReplaceAllString(arch, inner), nil
	case "inside":
		// Flattened arch: append before closing </view>
		i := strings.LastIndex(arch, "</view>")
		if i < 0 {
			return arch, fmt.Errorf("inherit xpath: no </view> in arch for position=inside")
		}
		return arch[:i] + inner + arch[i:], nil
	default:
		return arch, fmt.Errorf("unsupported xpath position %q", op.Position)
	}
}

func stripDataWrapper(s string) string {
	s = strings.TrimSpace(s)
	if m := dataWrapperRe.FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return s
}
