package web_test

import (
	"net/http/httptest"
	"sumeru/core/server/web"
	"testing"
)

func TestParseRelSearchRequest(t *testing.T) {
	req := httptest.NewRequest("GET", web.TestRelSearchRoute+"?model=core.partner&q=acme&limit=50&filter_field=country_id&filter_id=12", nil)
	request := web.ParseRelSearchRequest(req)

	if request.ModelName != "core.partner" || request.Query != "acme" || request.Limit != 50 {
		t.Fatalf("unexpected base fields: %+v", request)
	}
	if request.FilterField != "country_id" || request.FilterID != 12 {
		t.Fatalf("unexpected filter fields: %+v", request)
	}
}

func TestParseRelSearchRequestDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", web.TestRelSearchRoute+"?model=core.company", nil)
	request := web.ParseRelSearchRequest(req)

	if request.Query != "" || request.Limit != web.TestDefaultRelSearchLimit || request.FilterID != 0 {
		t.Fatalf("unexpected defaults: %+v", request)
	}
}

func TestQueryIntOrDefault(t *testing.T) {
	if got := web.QueryIntOrDefault("", 20); got != 20 {
		t.Fatalf("got %d want default 20", got)
	}
	if got := web.QueryIntOrDefault("bad", 20); got != 20 {
		t.Fatalf("invalid input should use default, got %d", got)
	}
	if got := web.QueryIntOrDefault("15", 20); got != 15 {
		t.Fatalf("got %d want 15", got)
	}
}

func TestQueryInt64OrZero(t *testing.T) {
	if got := web.QueryInt64OrZero(""); got != 0 {
		t.Fatalf("got %d want 0", got)
	}
	if got := web.QueryInt64OrZero("7"); got != 7 {
		t.Fatalf("got %d want 7", got)
	}
}
