package orm_test

import (
	"testing"

	"sumeru/core/orm"
)

func TestRewriteViewArchTreeToList(t *testing.T) {
	in := `<tree><field name="a"/></tree>`
	want := `<list><field name="a"/></list>`
	if got := orm.RewriteViewArchTreeToList(in); got != want {
		t.Fatalf("RewriteViewArchTreeToList() = %q; want %q", got, want)
	}
	in = `<view type="tree" model="m"><tree/></view>`
	want = `<view type="list" model="m"><list/></view>`
	if got := orm.RewriteViewArchTreeToList(in); got != want {
		t.Fatalf("RewriteViewArchTreeToList() = %q; want %q", got, want)
	}
}
