package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestParseChatterPostForm(t *testing.T) {
	body := "model=sale.order&res_id=42&body=Hello&next=%2Fweb%3Fid%3D42"
	req := httptest.NewRequest("POST", web.TestChatterPostRoute, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	form := web.ParseChatterPostForm(req)
	if form.ModelName != "sale.order" || form.RecordIDRaw != "42" || form.Body != "Hello" || form.Next != "/web?id=42" {
		t.Fatalf("unexpected form: %+v", form)
	}
}

func TestChatterBodyTooLong(t *testing.T) {
	shortBody := strings.Repeat("a", web.TestMaxChatterBodyRunes)
	if web.ChatterBodyTooLong(shortBody) {
		t.Fatal("body at limit should be allowed")
	}
	if !web.ChatterBodyTooLong(shortBody + "b") {
		t.Fatal("body over limit should be rejected")
	}
}

func TestParseChatterRecordID(t *testing.T) {
	recorder := httptest.NewRecorder()
	recordID, ok := web.ParseChatterRecordID(recorder, "15")
	if !ok || recordID != 15 {
		t.Fatalf("got id=%d ok=%v", recordID, ok)
	}

	recorder = httptest.NewRecorder()
	_, ok = web.ParseChatterRecordID(recorder, "0")
	if ok || recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for invalid id, got ok=%v status=%d", ok, recorder.Code)
	}
}
