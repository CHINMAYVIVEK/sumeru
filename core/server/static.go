package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/base"
	"sumeru/core/engine/render"
	"sumeru/core/server/config"
)

// registerBrandingAndStatic wires optional brand CSS, logo, and /static/ asset tree.
func registerBrandingAndStatic() {
	if config.AppConfig.BrandCSS != "" {
		if fi, err := os.Stat(config.AppConfig.BrandCSS); err == nil && !fi.IsDir() {
			base.SetExtraStylesheetURLs(base.SetExtraStylesheetURLsInput{URLs: []string{"/static/brand.css"}})
			http.HandleFunc("/static/brand.css", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, config.AppConfig.BrandCSS)
			})
			log.Printf("Brand stylesheet: %s → /static/brand.css", config.AppConfig.BrandCSS)
		} else {
			log.Printf("Warning: brand_css path not found or not a file: %s", config.AppConfig.BrandCSS)
		}
	}

	logoURL := ""
	if config.AppConfig.LogoPath != "" {
		if fi, err := os.Stat(config.AppConfig.LogoPath); err == nil && !fi.IsDir() {
			logoPath := config.AppConfig.LogoPath
			logoURL = "/static/app-logo"
			http.HandleFunc("/static/app-logo", func(w http.ResponseWriter, r *http.Request) {
				ext := strings.ToLower(filepath.Ext(logoPath))
				switch ext {
				case ".svg":
					w.Header().Set("Content-Type", "image/svg+xml")
				case ".png":
					w.Header().Set("Content-Type", "image/png")
				case ".jpg", ".jpeg":
					w.Header().Set("Content-Type", "image/jpeg")
				case ".webp":
					w.Header().Set("Content-Type", "image/webp")
				default:
					w.Header().Set("Content-Type", "application/octet-stream")
				}
				http.ServeFile(w, r, logoPath)
			})
			log.Printf("Logo: %s → /static/app-logo", logoPath)
		} else {
			log.Printf("Warning: logo_path not found or not a file: %s", config.AppConfig.LogoPath)
		}
	}

	render.SetShellBranding(render.ShellBranding{
		LogoURL: logoURL,
		Company: strings.TrimSpace(config.AppConfig.CompanyDisplayName),
		User:    strings.TrimSpace(config.AppConfig.UserDisplayName),
	})

	fs := http.FileServer(http.Dir(config.AppConfig.AssetsPath))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Printf("Serving static files from %s", filepath.Clean(config.AppConfig.AssetsPath))
}
