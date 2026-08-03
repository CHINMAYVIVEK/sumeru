package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/router"
)

func TestRegister_lastWinsAndMethodDispatch(t *testing.T) {
	router.Clear()
	t.Cleanup(router.Clear)
	router.RequireSession = nil
	router.ResolveUID = nil

	var first, second, postHit bool
	router.Register(http.MethodGet, "/t", router.AuthNone, func(w http.ResponseWriter, r *http.Request) {
		first = true
	})
	router.Register(http.MethodGet, "/t", router.AuthNone, func(w http.ResponseWriter, r *http.Request) {
		second = true
		w.WriteHeader(http.StatusOK)
	})
	router.Register(http.MethodPost, "/t", router.AuthNone, func(w http.ResponseWriter, r *http.Request) {
		postHit = true
		w.WriteHeader(http.StatusCreated)
	})

	mux := http.NewServeMux()
	router.Apply(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/t", nil))
	if first || !second {
		t.Fatalf("last register should win: first=%v second=%v", first, second)
	}

	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, httptest.NewRequest(http.MethodPost, "/t", nil))
	if !postHit || rr2.Code != http.StatusCreated {
		t.Fatalf("POST dispatch failed: hit=%v code=%d", postHit, rr2.Code)
	}
}

func TestAuthAPIKey_gate(t *testing.T) {
	router.Clear()
	t.Cleanup(func() {
		router.Clear()
		router.ResolveUID = nil
	})

	router.ResolveUID = func(r *http.Request) int { return 0 }
	router.Register(http.MethodGet, "/secure", router.AuthAPIKey, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux := http.NewServeMux()
	router.Apply(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/secure", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	router.ResolveUID = func(r *http.Request) int { return 7 }
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, httptest.NewRequest(http.MethodGet, "/secure", nil))
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 with uid, got %d", rr2.Code)
	}
}

func TestAuthSession_gate(t *testing.T) {
	router.Clear()
	t.Cleanup(func() {
		router.Clear()
		router.RequireSession = nil
	})

	router.RequireSession = func(w http.ResponseWriter, r *http.Request) bool {
		http.Error(w, "login", http.StatusFound)
		return false
	}
	router.Register(http.MethodGet, "/sess", router.AuthSession, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux := http.NewServeMux()
	router.Apply(mux)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/sess", nil))
	if rr.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", rr.Code)
	}
}
