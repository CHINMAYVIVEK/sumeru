package parser

import "testing"

func TestIsTruthyAttr(t *testing.T) {
	t.Parallel()
	truthy := []string{"1", "true", "TRUE", " yes ", "on"}
	for _, v := range truthy {
		if !IsTruthyAttr(v) {
			t.Errorf("IsTruthyAttr(%q) = false; want true", v)
		}
	}
	falsy := []string{"", "0", "false", "no", "off"}
	for _, v := range falsy {
		if IsTruthyAttr(v) {
			t.Errorf("IsTruthyAttr(%q) = true; want false", v)
		}
	}
}

func TestAttrLiteralOrExpr(t *testing.T) {
	t.Parallel()
	lit, truthy, expr := AttrLiteralOrExpr("1")
	if !lit || !truthy || expr != "" {
		t.Fatalf("true literal: %v %v %q", lit, truthy, expr)
	}
	lit, truthy, expr = AttrLiteralOrExpr("false")
	if !lit || truthy || expr != "" {
		t.Fatalf("false literal: %v %v %q", lit, truthy, expr)
	}
	lit, truthy, expr = AttrLiteralOrExpr("state != 'done'")
	if lit || truthy || expr != "state != 'done'" {
		t.Fatalf("expr: %v %v %q", lit, truthy, expr)
	}
}
