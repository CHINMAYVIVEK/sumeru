package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSmoke_unauthenticatedWebRedirectsToLogin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/web", WebHandler)
	h := SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("GET /web: status %d; want %d", rr.Code, http.StatusFound)
	}
	loc := rr.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil || !strings.HasPrefix(u.Path, "/web/login") {
		t.Fatalf("Location %q: want path prefix /web/login", loc)
	}
	if u.Query().Get("next") == "" {
		t.Fatalf("Location %q: want non-empty next query", loc)
	}
}

func TestSmoke_unauthenticatedRecordSaveRedirectsToLogin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/web/record/save", RecordSaveHandler)
	h := SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/web/record/save", strings.NewReader("model=core.user"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("POST /web/record/save: status %d; want %d", rr.Code, http.StatusFound)
	}
	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, "/web/login") {
		t.Fatalf("Location %q: want /web/login", loc)
	}
}

func TestSmoke_registerAppRoutesRootRedirectsToHome(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAppRoutes(mux)
	h := SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("GET /: status %d; want %d", rr.Code, http.StatusFound)
	}
	if g, w := rr.Header().Get("Location"), "/web/home"; g != w {
		t.Fatalf("Location %q; want %q", g, w)
	}
}

func TestSmoke_registerAppRoutesWebAppsTrailingSlashRedirect(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAppRoutes(mux)
	h := SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/apps/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("GET /web/apps/: status %d; want %d", rr.Code, http.StatusFound)
	}
	if g, w := rr.Header().Get("Location"), "/web/apps"; g != w {
		t.Fatalf("Location %q; want %q", g, w)
	}
}
