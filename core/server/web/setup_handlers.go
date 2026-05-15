package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

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
}

// SetupInitHandler runs database sync, installs base, bootstraps security from the JSON wizard payload, then restarts.
func SetupInitHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(request.Body, maxSetupInitBodyBytes+1))
	if err != nil {
		http.Error(responseWriter, "Could not read body", http.StatusBadRequest)
		return
	}
	if len(body) > maxSetupInitBodyBytes {
		http.Error(responseWriter, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	var payload setupInitRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(responseWriter, "Expected JSON with company_name, lang, admin_name, email, password", http.StatusBadRequest)
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
		log.Printf("Setup: RunFirstTimeInstallSync failed: %v", err)
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := module.InstallModuleByName(securityContext, "base"); err != nil {
		log.Printf("Setup: Install base failed: %v", err)
		http.Error(responseWriter, fmt.Sprintf("Install base failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := orm.EnsureBootstrapSecurityFromSetup(params); err != nil {
		log.Printf("Setup: Security bootstrap failed: %v", err)
		http.Error(responseWriter, fmt.Sprintf("Security bootstrap failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := orm.BackfillSysMenuModule(); err != nil {
		log.Printf("Warning: backfill sys.menu.module: %v", err)
	}

	go func() {
		time.Sleep(400 * time.Millisecond)
		log.Println("Setup: self-restarting server…")
		if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
			log.Fatalf("Setup: self-restart failed: %v", err)
		}
	}()
	fmt.Fprintln(responseWriter, "Setup complete — server is restarting…")
}

// SetupPageHandler renders the setup page from templates/setup.html.
func SetupPageHandler(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join(config.AppConfig.TemplatesPath, "setup.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Setup: parse %s: %v", tmplPath, err)
		http.Error(w, "Setup template missing", http.StatusInternalServerError)
		return
	}
	data := struct {
		DbName      string
		Stylesheets []string
	}{
		DbName:      config.AppConfig.DbName,
		Stylesheets: assets.DefaultStylesheetURLs(),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Setup: execute template: %v", err)
	}
}
