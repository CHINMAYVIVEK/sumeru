package web

import (
	"context"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sumeru/core/applog"
	"sumeru/core/engine/assets"
	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/sdk/platformmsg"
	"sumeru/core/server/config"

	"golang.org/x/crypto/bcrypt"
)

const invalidLoginMessage = "Invalid login or password."

// loginPageData is the view model for templates/login.html.
type loginPageData struct {
	Next        string
	Error       string
	Stylesheets []string
	LogoURL     string
}

// loginUser holds the fields needed to verify a password at sign-in.
type loginUser struct {
	ID           int
	PasswordHash string
	Active       bool
}

var (
	loginTemplateOnce sync.Once
	cachedLoginTmpl   *template.Template
	loginTemplateErr  error
)

// APIKeyUserID resolves X-API-Key or Authorization: Bearer credentials to a user id.
func APIKeyUserID(r *http.Request) int {
	if r == nil {
		return 0
	}
	raw := apiKeyFromRequest(r)
	if raw == "" {
		return 0
	}
	return orm.UIDFromAPIKey(r.Context(), raw)
}

// AuthenticatedUserID returns the session user id, or the API key user id when no session exists.
func AuthenticatedUserID(r *http.Request) int {
	if uid := SessionUserID(r); uid > 0 {
		return uid
	}
	return APIKeyUserID(r)
}

// SecurityMiddleware attaches request_id, authenticated uid, and active company to each request.
func SecurityMiddleware(next http.Handler) http.Handler {
	if next == nil {
		next = http.DefaultServeMux
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := requestIDFromHeader(r)
		w.Header().Set("X-Request-ID", requestID)

		ctx := enrichRequestContext(r, requestID)
		r = r.WithContext(ctx)

		logHTTPRequestStart(ctx, r)

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		logHTTPRequestEnd(ctx, r, recorder.status, time.Since(start))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// requireLogin redirects anonymous browser requests to the login page with a safe return URL.
func requireLogin(w http.ResponseWriter, r *http.Request) bool {
	if SessionUserID(r) > 0 {
		return true
	}
	returnTo := SafePathNext(r.URL.RequestURI(), homeRoute)
	http.Redirect(w, r, loginRoute+"?next="+url.QueryEscape(returnTo), http.StatusFound)
	return false
}

// LoginGet renders the login form for anonymous users.
func LoginGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if SessionUserID(r) > 0 {
		next := SafePathNext(r.URL.Query().Get("next"), homeRoute)
		http.Redirect(w, r, next, http.StatusFound)
		return
	}
	next := strings.TrimSpace(r.URL.Query().Get("next"))
	writeLoginPage(w, r, http.StatusOK, next, "")
}

// LoginPost validates credentials, opens a session, and redirects to the requested page.
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
	next := SafePathNext(r.PostFormValue("next"), homeRoute)
	ip := clientIP(r)

	user, err := lookupLoginUser(r.Context(), login)
	if err != nil || !userCanAuthenticate(user) {
		recordFailedLogin(r.Context(), 0, ip, "login="+login)
		writeLoginPage(w, r, http.StatusUnauthorized, next, invalidLoginMessage)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		recordFailedLogin(r.Context(), user.ID, ip, "bad password")
		writeLoginPage(w, r, http.StatusUnauthorized, next, invalidLoginMessage)
		return
	}
	if err := CreateSession(w, user.ID); err != nil {
		WebLogf(r.Context(), loginRoute, "session: %v", err)
		http.Error(w, "Could not start session", http.StatusInternalServerError)
		return
	}

	orm.AppendUserLog(r.Context(), user.ID, ip, "success")
	http.Redirect(w, r, next, http.StatusSeeOther)
}

// LogoutGet destroys the session cookie and returns the browser to the login page.
func LogoutGet(w http.ResponseWriter, r *http.Request) {
	DestroySession(w, r)
	http.Redirect(w, r, loginRoute, http.StatusFound)
}

// ActionResetPassword accepts a reset request from an authenticated user (email delivery not yet wired).
func ActionResetPassword(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}
	userID := strings.TrimSpace(r.PostFormValue("id"))
	login := strings.TrimSpace(r.PostFormValue("login"))
	WebLogf(r.Context(), resetPasswordRoute,
		"requested for user id=%s login=%q (email not yet wired)", userID, login)
	redirectWithWebMessage(w, r, r.PostFormValue("next"), "reset_requested")
}

func apiKeyFromRequest(r *http.Request) string {
	if key := strings.TrimSpace(r.Header.Get("X-API-Key")); key != "" {
		return key
	}
	return bearerToken(r.Header.Get("Authorization"))
}

func bearerToken(authHeader string) string {
	auth := strings.TrimSpace(authHeader)
	const prefix = "bearer "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(auth[len(prefix):])
}

func requestIDFromHeader(r *http.Request) string {
	if rid := strings.TrimSpace(r.Header.Get("X-Request-ID")); rid != "" {
		return rid
	}
	return applog.NewRequestID()
}

func enrichRequestContext(r *http.Request, requestID string) context.Context {
	ctx := applog.ContextWithRequestID(r.Context(), requestID)
	uid := AuthenticatedUserID(r)
	ctx = orm.ContextWithUID(ctx, uid)
	if uid > 0 {
		ctx = orm.ContextWithCompanyID(ctx, orm.ActiveCompanyIDForUser(ctx, uid))
	}
	return ctx
}

func logHTTPRequestStart(ctx context.Context, r *http.Request) {
	applog.Debug(ctx, applog.Event{
		Message:   "HTTP request started",
		Component: "web",
		Operation: "request",
		Status:    "success",
		Context: map[string]interface{}{
			"route":  r.URL.Path,
			"method": r.Method,
		},
	})
}

func logHTTPRequestEnd(ctx context.Context, r *http.Request, status int, duration time.Duration) {
	ev := applog.Event{
		Component: "web",
		Operation: "request",
		Duration:  duration,
		Context: map[string]interface{}{
			"route":       r.URL.Path,
			"method":      r.Method,
			"status_code": status,
		},
	}
	if status >= 500 {
		ev.Message = "HTTP request failed"
		ev.Status = "failure"
		applog.Error(ctx, ev)
		return
	}
	ev.Message = "HTTP request completed"
	ev.Status = "success"
	applog.Debug(ctx, ev)
}

func getLoginTemplate() (*template.Template, error) {
	loginTemplateOnce.Do(func() {
		path := filepath.Join(config.AppConfig.TemplatesPath, "login.html")
		cachedLoginTmpl, loginTemplateErr = template.ParseFiles(path)
	})
	return cachedLoginTmpl, loginTemplateErr
}

func newLoginPageData(next, errMsg string) loginPageData {
	return loginPageData{
		Next:        next,
		Error:       errMsg,
		Stylesheets: assets.LoginStylesheetURLs(),
		LogoURL:     render.ShellLogoURL(),
	}
}

func writeLoginPage(w http.ResponseWriter, r *http.Request, status int, next, errMsg string) {
	tmpl, err := getLoginTemplate()
	if err != nil {
		if status == http.StatusOK {
			WebLogf(r.Context(), loginRoute, "%s: login template: %v", platformmsg.MsgHTTPTemplateError, err)
			http.Error(w, "Login page unavailable", http.StatusInternalServerError)
			return
		}
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	_ = tmpl.Execute(w, newLoginPageData(next, errMsg))
}

func lookupLoginUser(ctx context.Context, login string) (loginUser, error) {
	tbl := orm.MustQuotedTableName("core.user")
	var user loginUser
	err := orm.DB.QueryRowContext(ctx,
		`SELECT id, COALESCE(password, ''), active FROM `+tbl+` WHERE LOWER(TRIM(login)) = LOWER(TRIM($1)) LIMIT 1`,
		login,
	).Scan(&user.ID, &user.PasswordHash, &user.Active)
	return user, err
}

func userCanAuthenticate(user loginUser) bool {
	return user.Active && strings.TrimSpace(user.PasswordHash) != ""
}

func recordFailedLogin(ctx context.Context, userID int, ip, auditNote string) {
	orm.AppendUserLog(ctx, userID, ip, "failure")
	orm.AppendAudit(ctx, "login_fail", "core.user", int64(userID), nil, nil, auditNote)
}
