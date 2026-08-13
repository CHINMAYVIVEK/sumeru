package server

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/engine/assets"
	"sumeru/core/engine/render"
	"sumeru/core/module"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// registerBrandingAndStatic wires core CSS (theme + layout slices), optional addon theme overrides,
// optional brand CSS, logo, and the /static/ asset tree.
func registerBrandingAndStatic() {
	ctx := context.Background()
	extraCSS := append([]string(nil), assets.DefaultStylesheetURLs()...)
	if addonModuleInstalled("sumeru_ai") {
		withAI := make([]string, 0, len(extraCSS)+1)
		withAI = append(withAI, extraCSS[:6]...)
		withAI = append(withAI, assets.AIStylesheetURL())
		withAI = append(withAI, extraCSS[6:]...)
		extraCSS = withAI
	}

	addonNames := make([]string, 0, len(module.LoadedAddons))
	for name := range module.LoadedAddons {
		addonNames = append(addonNames, name)
	}
	sort.Strings(addonNames)
	for _, name := range addonNames {
		a := module.LoadedAddons[name]
		if a == nil {
			continue
		}
		if !addonModuleInstalled(name) {
			continue
		}
		overridePath := filepath.Join(a.Path, "static/css/theme-overrides.css")
		if fi, err := os.Stat(overridePath); err != nil || fi.IsDir() {
			continue
		}
		urlPath := "/static/addon-css/" + name + ".css"
		p := overridePath
		modName := name
		http.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
			if !addonModuleInstalled(modName) {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			http.ServeFile(w, r, p)
		})
		extraCSS = append(extraCSS, urlPath)
		applog.InfoMsg(ctx, "web", "static", "Registered addon theme overrides",
			map[string]interface{}{"path": overridePath, "url": urlPath})
	}

	if config.AppConfig.BrandCSS != "" {
		if fileInfo, err := os.Stat(config.AppConfig.BrandCSS); err == nil && !fileInfo.IsDir() {
			extraCSS = append(extraCSS, "/static/brand.css")
			http.HandleFunc("/static/brand.css", func(responseWriter http.ResponseWriter, request *http.Request) {
				http.ServeFile(responseWriter, request, config.AppConfig.BrandCSS)
			})
			applog.InfoMsg(ctx, "web", "static", "Registered brand stylesheet",
				map[string]interface{}{"path": config.AppConfig.BrandCSS})
		} else {
			applog.WarnMsg(ctx, "web", "static", "brand_css path not found or not a file", nil,
				map[string]interface{}{"path": config.AppConfig.BrandCSS})
		}
	}

	render.SetExtraStylesheetURLs(extraCSS)

	logoURL := ""
	if config.AppConfig.LogoPath != "" {
		if fileInfo, err := os.Stat(config.AppConfig.LogoPath); err == nil && !fileInfo.IsDir() {
			logoPath := config.AppConfig.LogoPath
			logoURL = "/static/app-logo"
			http.HandleFunc("/static/app-logo", func(responseWriter http.ResponseWriter, request *http.Request) {
				extension := strings.ToLower(filepath.Ext(logoPath))
				switch extension {
				case ".svg":
					responseWriter.Header().Set("Content-Type", "image/svg+xml")
				case ".png":
					responseWriter.Header().Set("Content-Type", "image/png")
				case ".jpg", ".jpeg":
					responseWriter.Header().Set("Content-Type", "image/jpeg")
				case ".webp":
					responseWriter.Header().Set("Content-Type", "image/webp")
				default:
					responseWriter.Header().Set("Content-Type", "application/octet-stream")
				}
				http.ServeFile(responseWriter, request, logoPath)
			})
			applog.InfoMsg(ctx, "web", "static", "Registered application logo",
				map[string]interface{}{"path": logoPath})
		} else {
			applog.WarnMsg(ctx, "web", "static", "logo_path not found or not a file", nil,
				map[string]interface{}{"path": config.AppConfig.LogoPath})
		}
	}

	render.SetShellBranding(render.ShellBranding{
		LogoURL: logoURL,
		Company: strings.TrimSpace(config.AppConfig.CompanyDisplayName),
		User:    strings.TrimSpace(config.AppConfig.UserDisplayName),
	})

	fileServer := http.FileServer(http.Dir(config.AppConfig.AssetsPath))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))
	applog.InfoMsg(ctx, "web", "static", "Serving static files",
		map[string]interface{}{"path": filepath.Clean(config.AppConfig.AssetsPath)})
}

func addonModuleInstalled(technicalName string) bool {
	if orm.DB == nil {
		return false
	}
	tbl := orm.MustQuotedTableName("sys.module")
	var state string
	err := orm.DB.QueryRow(`SELECT state FROM `+tbl+` WHERE name = $1 AND active = true`, strings.TrimSpace(technicalName)).Scan(&state)
	return err == nil && strings.TrimSpace(state) == "installed"
}
