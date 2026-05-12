package viewinherit

import (
	"strings"
	"testing"
)

func TestApplyInheritArchAfter(t *testing.T) {
	parent := `<view model="sale.order" type="tree"><field name="a" string="A"/><field name="b" string="B"/></view>`
	frag := `<xpath expr="//field[@name='a']" position="after"><field name="z" string="Z"/></xpath>`
	out, err := ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(out, `name="a"`, `name="z"`, `name="b"`) {
		t.Fatalf("unexpected merge: %s", out)
	}
}

func containsInOrder(s string, parts ...string) bool {
	pos := 0
	for _, p := range parts {
		i := indexFrom(s, p, pos)
		if i < 0 {
			return false
		}
		pos = i + len(p)
	}
	return true
}

func indexFrom(s, sub string, start int) int {
	if start > len(s) {
		return -1
	}
	idx := strings.Index(s[start:], sub)
	if idx < 0 {
		return -1
	}
	return start + idx
}
