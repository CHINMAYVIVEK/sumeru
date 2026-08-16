package web_test

import (
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"

	"sumeru/core/orm"
)

func TestParseRecordSaveForm(t *testing.T) {
	body := "model=core.company&id=12&next=%2Fweb%3Fmenu_id%3D1"
	req := httptest.NewRequest("POST", web.TestRecordSaveRoute, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	form := web.ParseRecordSaveForm(req)
	if form.ModelName != "core.company" || form.RecordIDRaw != "12" || form.Next != "/web?menu_id=1" {
		t.Fatalf("unexpected form: %+v", form)
	}
}

func TestIsNewRecord(t *testing.T) {
	if !web.IsNewRecord("") || !web.IsNewRecord("0") {
		t.Fatal("empty or zero id should be treated as new record")
	}
	if web.IsNewRecord("42") {
		t.Fatal("non-zero id should not be treated as new record")
	}
}

func TestWorkspaceFormURL(t *testing.T) {
	got := web.WorkspaceFormURL("/web?action=5&menu_id=2&edit=1", 99)
	assertQueryContains(t, got, map[string]string{
		web.TestWorkspaceActionParam:   "5",
		web.TestWorkspaceMenuIDParam:   "2",
		web.TestWorkspaceViewTypeParam: web.TestWorkspaceViewModeForm,
		web.TestWorkspaceRecordIDParam: "99",
	})
	if strings.Contains(got, "edit=") {
		t.Fatalf("edit param should be removed: %q", got)
	}

	fallback := web.WorkspaceFormURL("://bad", 7)
	assertQueryContains(t, fallback, map[string]string{
		web.TestWorkspaceViewTypeParam: web.TestWorkspaceViewModeForm,
		web.TestWorkspaceRecordIDParam: "7",
	})
}

func TestIsSkippedSaveFormField(t *testing.T) {
	if !web.IsSkippedSaveFormField(web.TestRecordModelField) || !web.IsSkippedSaveFormField(web.TestNextField) {
		t.Fatal("expected meta fields to be skipped")
	}
	if web.IsSkippedSaveFormField("name") {
		t.Fatal("model data fields should not be skipped")
	}
}

func TestCoerceSaveFieldValue(t *testing.T) {
	value, skip, err := web.CoerceSaveFieldValue("active", orm.Boolean, "true")
	if err != nil || skip || value != true {
		t.Fatalf("got %#v skip=%v err=%v", value, skip, err)
	}

	value, skip, err = web.CoerceSaveFieldValue("amount", orm.Float, "3.5")
	if err != nil || skip || value != 3.5 {
		t.Fatalf("got %#v skip=%v err=%v", value, skip, err)
	}

	_, skip, err = web.CoerceSaveFieldValue("partner_id", orm.Many2One, "")
	if err != nil || !skip {
		t.Fatalf("empty many2one should skip: skip=%v err=%v", skip, err)
	}

	value, skip, err = web.CoerceSaveFieldValue("date_deadline", orm.Date, "")
	if err != nil || skip || value != nil {
		t.Fatalf("empty date should coerce to nil: value=%#v skip=%v err=%v", value, skip, err)
	}

	value, skip, err = web.CoerceSaveFieldValue("date_last_stage_update", orm.DateTime, "2026-08-15T14:30")
	if err != nil || skip || value != "2026-08-15T14:30" {
		t.Fatalf("datetime value=%#v skip=%v err=%v", value, skip, err)
	}
}
