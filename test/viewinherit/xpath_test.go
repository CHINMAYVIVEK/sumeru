package viewinherit_test

import (
	"strings"
	"testing"

	"sumeru/core/engine/viewinherit"
)

func TestApplyInheritArchAfter(t *testing.T) {
	parent := `<view model="sale.order" type="list"><field name="a" string="A"/><field name="b" string="B"/></view>`
	frag := `<xpath expr="//field[@name='a']" position="after"><field name="z" string="Z"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(out, `name="a"`, `name="z"`, `name="b"`) {
		t.Fatalf("unexpected merge: %s", out)
	}
}

func TestApplyInheritArchAfterMarshaledField(t *testing.T) {
	parent := `<view id="" model="sale.order" type="list" title="" open=""><field name="state" string="Status" widget="" placeholder="" options=""></field><field name="amount" string="Total" widget="" placeholder="" options=""></field></view>`
	frag := `<xpath expr="//field[@name='state']" position="after"><field name="phone" string="Phone"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(out, `name="state"`, `name="phone"`, `name="amount"`) {
		t.Fatalf("unexpected merge: %s", out)
	}
}

func TestApplyInheritArchButtonAndAttributes(t *testing.T) {
	parent := `<view type="form"><header><button name="action_lost" string="Lost" type="object"></button></header><field name="phone" string="Phone"></field></view>`
	frag := `<xpath expr="//button[@name='action_lost']" position="attributes"><attribute name="string">Mark Lost</attribute></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `string="Mark Lost"`) {
		t.Fatalf("button attr: %s", out)
	}
	frag2 := `<xpath expr="//field[@name='phone']" position="attributes"><attribute name="invisible">1</attribute></xpath>`
	out2, err := viewinherit.ApplyInheritArch(out, frag2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out2, `invisible="1"`) {
		t.Fatalf("field attr: %s", out2)
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
