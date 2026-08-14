package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"sumeru/core/applog"
	"sumeru/core/engine/assets"
	"sumeru/core/module"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const maxSetupInitBodyBytes = 1 << 17 // 128 KiB

type setupInitRequest struct {
	CompanyName string `json:"company_name"`
	Lang        string `json:"lang"`
	AdminName   string `json:"admin_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	SetupToken  string `json:"setup_token"`
}

// SetupInitHandler runs database sync, installs base, bootstraps security from the JSON wizard payload, then restarts.
func SetupInitHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !RequirePOST(w, r) {
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxSetupInitBodyBytes+1))
	if err != nil {
		http.Error(w, "Could not read body", http.StatusBadRequest)
		return
	}
	if len(body) > maxSetupInitBodyBytes {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	var payload setupInitRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Expected JSON with company_name, lang, admin_name, email, password", http.StatusBadRequest)
		return
	}
	if !allowSetupRequest(w, r, payload.SetupToken) {
		return
	}

	params := orm.SetupAdminParams{
		CompanyName: payload.CompanyName,
		Lang:        payload.Lang,
		FullName:    payload.AdminName,
		Email:       payload.Email,
		Password:    payload.Password,
	}

	securityContext := orm.ContextWithBypass(context.Background(), true)

	if err := module.RunFirstTimeInstallSync(securityContext); err != nil {
		applog.Error(ctx, applog.Event{
			Message: "First-time install sync failed", Component: "web", Operation: "setup",
			Status: "failure", Err: err,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := module.InstallModuleByName(securityContext, "base"); err != nil {
		applog.Error(ctx, applog.Event{
			Message: "Install base module failed", Component: "web", Operation: "setup",
			Status: "failure", Err: err,
		})
		http.Error(w, fmt.Sprintf("Install base failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := orm.EnsureBootstrapSecurityFromSetup(params); err != nil {
		applog.Error(ctx, applog.Event{
			Message: "Security bootstrap failed", Component: "web", Operation: "setup",
			Status: "failure", Err: err,
		})
		http.Error(w, fmt.Sprintf("Security bootstrap failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := orm.RunMenuDataFixes(); err != nil {
		applog.WarnMsg(ctx, "web", "setup", "Menu data fixes reported issues", err, nil)
	}

	go func() {
		time.Sleep(400 * time.Millisecond)
		applog.InfoMsg(context.Background(), "web", "setup", "Self-restarting server after setup", nil)
		if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
			applog.Fatal(context.Background(), "Setup self-restart failed", "err", err)
		}
	}()
	fmt.Fprintln(w, "Setup complete — server is restarting…")
}

// SetupPageHandler renders the setup page from templates/setup.html.
func SetupPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if orm.IsInitialized() {
		http.Error(w, "Setup already completed", http.StatusForbidden)
		return
	}
	if config.AppConfig.SetupLocalhostOnly && !isLoopbackIP(clientIP(r)) {
		http.Error(w, "Setup is restricted to localhost", http.StatusForbidden)
		return
	}
	tmplPath := filepath.Join(config.AppConfig.TemplatesPath, "setup.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		applog.Error(ctx, applog.Event{
			Message: "Failed to parse setup template", Component: "web", Operation: "setup",
			Status: "failure", Err: err, Context: map[string]interface{}{"template": tmplPath},
		})
		http.Error(w, "Setup template missing", http.StatusInternalServerError)
		return
	}
	data := struct {
		DbName             string
		Stylesheets        []string
		SetupTokenRequired bool
	}{
		DbName:             config.AppConfig.DbName,
		Stylesheets:        assets.LoginStylesheetURLs(),
		SetupTokenRequired: strings.TrimSpace(config.AppConfig.SetupToken) != "",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		applog.Error(ctx, applog.Event{
			Message: "Failed to execute setup template", Component: "web", Operation: "setup",
			Status: "failure", Err: err,
		})
	}
}
