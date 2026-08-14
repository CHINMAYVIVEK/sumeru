package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestParseRecordDeleteRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/web/record/delete?action=10&menu_id=3", strings.NewReader("model=sale.order&id=99"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	request, ok := web.ParseRecordDeleteRequest(recorder, req)
	if !ok {
		t.Fatalf("unexpected failure status=%d", recorder.Code)
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

	recorder := httptest.NewRecorder()
	_, ok := web.ParseRecordDeleteRequest(recorder, req)
	if ok || recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got ok=%v status=%d", ok, recorder.Code)
	}
}

func TestParseRecordDeleteRequestQueryFallback(t *testing.T) {
	req := httptest.NewRequest("POST", "/web/record/delete?model=core.company&id=7", nil)
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	request, ok := web.ParseRecordDeleteRequest(recorder, req)
	if !ok {
		t.Fatal("expected success from query params")
	}
	if request.ModelName != "core.company" || request.RecordID != 7 {
		t.Fatalf("unexpected request: %+v", request)
	}
}
