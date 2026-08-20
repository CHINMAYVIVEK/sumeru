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
