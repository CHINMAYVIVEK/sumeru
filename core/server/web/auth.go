package web

import (
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"sumeru/core/sdk/platformmsg"
	"sumeru/core/engine/assets"
	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/server/config"

	"golang.org/x/crypto/bcrypt"
)

// APIKeyUserID extracts a bearer / X-API-Key credential and resolves it to a user id.
func APIKeyUserID(r *http.Request) int {
	if r == nil {
		return 0
	}
	raw := strings.TrimSpace(r.Header.Get("X-API-Key"))
	if raw == "" {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			raw = strings.TrimSpace(auth[7:])
		}
	}
	if raw == "" {
		return 0
	}
	return orm.UIDFromAPIKey(r.Context(), raw)
}

// AuthenticatedUserID returns session uid, or API key uid if no session.
func AuthenticatedUserID(r *http.Request) int {
	if uid := SessionUserID(r); uid > 0 {
		return uid
	}
	return APIKeyUserID(r)
}

// SecurityMiddleware attaches orm.UIDFromContext from session cookie or API key.
func SecurityMiddleware(next http.Handler) http.Handler {
	if next == nil {
		next = http.DefaultServeMux
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := AuthenticatedUserID(r)
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
		q = "/web/home"
	}
	http.Redirect(w, r, "/web/login?next="+url.QueryEscape(q), http.StatusFound)
	return false
}

type loginTemplateData struct {
	Next        string
	Error       string
	Stylesheets []string
	LogoURL     string
}

// LoginGet renders the login form.
func LoginGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if SessionUserID(r) > 0 {
		next := SafePathNext(r.URL.Query().Get("next"), "/web/home")
		http.Redirect(w, r, next, http.StatusFound)
		return
	}
	tmplPath := filepath.Join(config.AppConfig.TemplatesPath, "login.html")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		WebLogf(r.Context(), "/web/login", "%s: login template: %v", platformmsg.MsgHTTPTemplateError, err)
		http.Error(w, "Login page unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = t.Execute(w, loginTemplateData{
		Next:        strings.TrimSpace(r.URL.Query().Get("next")),
		Error:       "",
		Stylesheets: assets.DefaultStylesheetURLs(),
		LogoURL:     render.ShellLogoURL(),
	})
}

// LoginPost authenticates and creates a session.
func LoginPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !ParsePostForm(w, r) {
		return
	}
	login := strings.TrimSpace(r.PostFormValue("login"))
	password := r.PostFormValue("password")
	next := SafePathNext(r.PostFormValue("next"), "/web/home")
	tbl := orm.GetTableName("core.user")
	var id int
	var hash string
	var active bool
	err := orm.DB.QueryRow(
		`SELECT id, COALESCE(password, ''), active FROM `+tbl+` WHERE LOWER(TRIM(login)) = LOWER(TRIM($1)) LIMIT 1`,
		login,
	).Scan(&id, &hash, &active)
	clientIP := r.RemoteAddr
	if err != nil || !active || strings.TrimSpace(hash) == "" {
		orm.AppendUserLog(r.Context(), 0, clientIP, "failure")
		orm.AppendAudit(r.Context(), "login_fail", "core.user", 0, nil, nil, "login="+login)
		renderLoginError(w, r, next, "Invalid login or password.")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		orm.AppendUserLog(r.Context(), id, clientIP, "failure")
		orm.AppendAudit(r.Context(), "login_fail", "core.user", int64(id), nil, nil, "bad password")
		renderLoginError(w, r, next, "Invalid login or password.")
		return
	}
	if err := CreateSession(w, id); err != nil {
		WebLogf(r.Context(), "/web/login", "session: %v", err)
		http.Error(w, "Could not start session", http.StatusInternalServerError)
		return
	}
	orm.AppendUserLog(r.Context(), id, clientIP, "success")
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
	_ = t.Execute(w, loginTemplateData{
		Next:        next,
		Error:       msg,
		Stylesheets: assets.DefaultStylesheetURLs(),
		LogoURL:     render.ShellLogoURL(),
	})
}

// LogoutGet destroys the session.
func LogoutGet(w http.ResponseWriter, r *http.Request) {
	DestroySession(w, r)
	http.Redirect(w, r, "/web/login", http.StatusFound)
}

// ActionResetPassword is a stub for the "Send reset instructions" button on the user form.
// WIP: email delivery is not yet wired; this handler logs the attempt and redirects.
func ActionResetPassword(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !RequirePOST(w, r) {
		return
	}
	if !ParsePostForm(w, r) {
		return
	}
	userID := strings.TrimSpace(r.PostFormValue("id"))
	login := strings.TrimSpace(r.PostFormValue("login"))
	WebLogf(r.Context(), "/web/action/reset_password", "requested for user id=%s login=%q (email not yet wired)", userID, login)
	// Redirect back with a flash-style query param so the UI can surface it.
	next := SafeWebNext(r.PostFormValue("next"), "/web/home")
	http.Redirect(w, r, next+"&msg=reset_requested", http.StatusSeeOther)
}
