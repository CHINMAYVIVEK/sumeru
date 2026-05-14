package web

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"sumeru/core/base/platformmsg"
	"sumeru/core/orm"
	"sumeru/core/server/config"

	"golang.org/x/crypto/bcrypt"
)

// SecurityMiddleware attaches orm.UIDFromContext from the session cookie for each request context.
func SecurityMiddleware(next http.Handler) http.Handler {
	if next == nil {
		next = http.DefaultServeMux
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := SessionUserID(r)
		ctx := orm.ContextWithUID(r.Context(), uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requireLogin(w http.ResponseWriter, r *http.Request) bool {
	if SessionUserID(r) > 0 {
		return true
	}
	q := r.URL.RequestURI()
	if q == "" {
		q = "/web"
	}
	http.Redirect(w, r, "/web/login?next="+url.QueryEscape(q), http.StatusFound)
	return false
}

// LoginGet renders the login form.
func LoginGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if SessionUserID(r) > 0 {
		next := strings.TrimSpace(r.URL.Query().Get("next"))
		if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
			next = "/web/apps"
		}
		http.Redirect(w, r, next, http.StatusFound)
		return
	}
	tmplPath := filepath.Join(config.AppConfig.TemplatesPath, "login.html")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("%s: login template: %v", platformmsg.MsgHTTPTemplateError, err)
		http.Error(w, "Login page unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = t.Execute(w, map[string]string{
		"Next":  strings.TrimSpace(r.URL.Query().Get("next")),
		"Error": "",
	})
}

// LoginPost authenticates and creates a session.
func LoginPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad form", http.StatusBadRequest)
		return
	}
	login := strings.TrimSpace(r.PostFormValue("login"))
	password := r.PostFormValue("password")
	next := strings.TrimSpace(r.PostFormValue("next"))
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		next = "/web/apps"
	}
	tbl := orm.GetTableName("core.user")
	var id int
	var hash string
	var active bool
	err := orm.DB.QueryRow(
		`SELECT id, COALESCE(password, ''), active FROM `+tbl+` WHERE LOWER(TRIM(login)) = LOWER(TRIM($1)) LIMIT 1`,
		login,
	).Scan(&id, &hash, &active)
	if err != nil || !active || strings.TrimSpace(hash) == "" {
		renderLoginError(w, r, next, "Invalid login or password.")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		renderLoginError(w, r, next, "Invalid login or password.")
		return
	}
	if err := CreateSession(w, id); err != nil {
		log.Printf("web: session: %v", err)
		http.Error(w, "Could not start session", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func renderLoginError(w http.ResponseWriter, r *http.Request, next, msg string) {
	tmplPath := filepath.Join(config.AppConfig.TemplatesPath, "login.html")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_ = t.Execute(w, map[string]string{
		"Next":  next,
		"Error": msg,
	})
}

// LogoutGet destroys the session.
func LogoutGet(w http.ResponseWriter, r *http.Request) {
	DestroySession(w, r)
	http.Redirect(w, r, "/web/login", http.StatusFound)
}

// ActionResetPassword is a stub for the "Send reset instructions" button on the user form.
// Email delivery is not yet wired; this handler logs the attempt and returns a clear message
// so the button doesn't silently fail. Replace with a real email implementation later.
func ActionResetPassword(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	userID := strings.TrimSpace(r.PostFormValue("id"))
	login := strings.TrimSpace(r.PostFormValue("login"))
	log.Printf("action_reset_password: requested for user id=%s login=%q (email not yet wired)", userID, login)
	// Redirect back with a flash-style query param so the UI can surface it.
	next := strings.TrimSpace(r.PostFormValue("next"))
	if next == "" || !strings.HasPrefix(next, "/web") || strings.HasPrefix(next, "//") {
		next = "/web"
	}
	http.Redirect(w, r, next+"&msg=reset_requested", http.StatusSeeOther)
}
