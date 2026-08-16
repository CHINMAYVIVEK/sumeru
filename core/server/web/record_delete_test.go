package web_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/web"
)

func TestParseRecordDeleteRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/web/record/delete?action=10&menu_id=3", strings.NewReader("model=sale.order&id=99"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	request, err := web.ParseRecordDeleteRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if request.ModelName != "sale.order" || request.RecordID != 99 || request.ActionID != "10" || request.MenuID != "3" {
		t.Fatalf("unexpected request: %+v", request)
	}
}

func TestParseRecordDeleteRequestMissingFields(t *testing.T) {
	req := httptest.NewRequest("POST", web.TestRecordDeleteRoute, strings.NewReader("model=sale.order"))
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	_, err := web.ParseRecordDeleteRequest(req)
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestParseRecordDeleteRequestQueryFallback(t *testing.T) {
	req := httptest.NewRequest("POST", "/web/record/delete?model=core.company&id=7", nil)
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	request, err := web.ParseRecordDeleteRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if request.ModelName != "core.company" || request.RecordID != 7 {
		t.Fatalf("unexpected request: %+v", request)
	}
}
