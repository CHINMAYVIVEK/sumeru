package viewinherit

import (
	"fmt"
	"regexp"
	"strings"
)

type xpathOp struct {
	Expr     string
	Position string
	Inner    string
}

var xpathBlockRe = regexp.MustCompile(`(?s)<xpath\s+expr="([^"]+)"\s+position="([^"]+)"\s*>(.*?)</xpath>`)
var xpathBlockReSingle = regexp.MustCompile(`(?s)<xpath\s+expr='([^']+)'\s+position='([^']+)'\s*>(.*?)</xpath>`)
var fieldNameFromExpr = regexp.MustCompile(`@name=['"]([^'"]+)['"]`)
var xpathTargetRe = regexp.MustCompile(`//(field|button)\[@name=['"]([^'"]+)['"]\]`)
var dataWrapperRe = regexp.MustCompile(`(?s)^\s*<data[^>]*>(.*)</data>\s*$`)
var attributeOpRe = regexp.MustCompile(`(?s)<attribute\s+name=['"]([^'"]+)['"]\s*>(.*?)</attribute>`)

// ApplyInheritArch parses <xpath> blocks in inheritFragment and applies them to parentArch.
// Supported: position after|before|inside|replace|attributes on //field[@name='x'] or //button[@name='x'].
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

func xpathTargetFromExpr(expr string) (tag, name string, err error) {
	expr = strings.TrimSpace(expr)
	if m := xpathTargetRe.FindStringSubmatch(expr); len(m) >= 3 {
		return strings.ToLower(m[1]), m[2], nil
	}
	m := fieldNameFromExpr.FindStringSubmatch(expr)
	if len(m) < 2 {
		return "", "", fmt.Errorf("unsupported xpath expr (need //field[@name='…'] or //button[@name='…']): %q", expr)
	}
	return "field", m[1], nil
}

func elementTagRe(tag, name string) *regexp.Regexp {
	q := regexp.QuoteMeta(name)
	return regexp.MustCompile(`<` + tag + `\s+[^>]*\bname="` + q + `"[^>]*(?:/>|>\s*</` + tag + `>)`)
}

func openingTagRe(tag, name string) *regexp.Regexp {
	q := regexp.QuoteMeta(name)
	return regexp.MustCompile(`<` + tag + `\s+[^>]*\bname="` + q + `"[^>]*>`)
}

func applyOne(arch string, op xpathOp) (string, error) {
	tag, name, err := xpathTargetFromExpr(strings.TrimSpace(op.Expr))
	if err != nil {
		return arch, err
	}
	pos := strings.ToLower(strings.TrimSpace(op.Position))
	inner := strings.TrimSpace(op.Inner)
	re := elementTagRe(tag, name)

	switch pos {
	case "after":
		loc := re.FindStringIndex(arch)
		if loc == nil {
			return arch, fmt.Errorf("inherit xpath: %s %q not found for position=after", tag, name)
		}
		insert := inner
		if !strings.HasPrefix(strings.TrimSpace(insert), "<") {
			insert = "<field name=\"" + insert + "\"/>"
		}
		return arch[:loc[1]] + insert + arch[loc[1]:], nil
	case "before":
		loc := re.FindStringIndex(arch)
		if loc == nil {
			return arch, fmt.Errorf("inherit xpath: %s %q not found for position=before", tag, name)
		}
		return arch[:loc[0]] + inner + arch[loc[0]:], nil
	case "replace":
		if !re.MatchString(arch) {
			return arch, fmt.Errorf("inherit xpath: %s %q not found for position=replace", tag, name)
		}
		return re.ReplaceAllString(arch, inner), nil
	case "inside":
		i := strings.LastIndex(arch, "</view>")
		if i < 0 {
			return arch, fmt.Errorf("inherit xpath: no </view> in arch for position=inside")
		}
		return arch[:i] + inner + arch[i:], nil
	case "attributes":
		return applyAttributes(arch, tag, name, inner)
	default:
		return arch, fmt.Errorf("unsupported xpath position %q", op.Position)
	}
}

func applyAttributes(arch, tag, name, inner string) (string, error) {
	re := openingTagRe(tag, name)
	loc := re.FindStringIndex(arch)
	if loc == nil {
		return arch, fmt.Errorf("inherit xpath: %s %q not found for position=attributes", tag, name)
	}
	old := arch[loc[0]:loc[1]]
	attrs := parseAttributeOps(inner)
	if len(attrs) == 0 {
		return arch, fmt.Errorf("inherit xpath: no <attribute> elements for position=attributes")
	}
	next := old
	for attrName, attrVal := range attrs {
		next = upsertXMLAttr(next, attrName, attrVal)
	}
	return arch[:loc[0]] + next + arch[loc[1]:], nil
}

func parseAttributeOps(inner string) map[string]string {
	out := map[string]string{}
	for _, m := range attributeOpRe.FindAllStringSubmatch(inner, -1) {
		name := strings.TrimSpace(m[1])
		if name == "" {
			continue
		}
		out[name] = strings.TrimSpace(m[2])
	}
	return out
}

func upsertXMLAttr(openTag, name, value string) string {
	nameRe := regexp.MustCompile(`\s` + regexp.QuoteMeta(name) + `="[^"]*"`)
	replacement := ` ` + name + `="` + value + `"`
	if nameRe.MatchString(openTag) {
		return nameRe.ReplaceAllString(openTag, replacement)
	}
	if strings.HasSuffix(openTag, "/>") {
		return strings.TrimSuffix(openTag, "/>") + replacement + `/>`
	}
	if strings.HasSuffix(openTag, ">") {
		return strings.TrimSuffix(openTag, ">") + replacement + `>`
	}
	return openTag + replacement
}

func stripDataWrapper(s string) string {
	s = strings.TrimSpace(s)
	if m := dataWrapperRe.FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return s
}
