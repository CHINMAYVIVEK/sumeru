package config

import (
	"os"
	"path/filepath"
	"strings"
)

// AbsPaths resolves relative config paths to absolute paths.
// AddonPaths is set from addons_path segments (platform addons live under addons/,
// e.g. addons/base). When sumeru_home is set, default assets/templates resolve
// under the standard sumeru checkout.
// Relative addons_path, assets_path, templates_path, etc. resolve against ConfigFileDir (INI directory).
func AbsPaths() error {
	su := strings.TrimSpace(AppConfig.SumeruHome)
	var sumeruAbs string
	var err error
	if su != "" {
		if !filepath.IsAbs(su) {
			su = filepath.Join(ConfigFileDir, su)
		}
		sumeruAbs, err = filepath.Abs(su)
		if err != nil {
			return err
		}
	}

	if strings.TrimSpace(AppConfig.AssetsPath) == "" {
		if sumeruAbs != "" {
			AppConfig.AssetsPath = filepath.Join(sumeruAbs, segCore, segEngine, segAssets)
		} else {
			AppConfig.AssetsPath = relPathDefaultAssets
		}
	}
	if strings.TrimSpace(AppConfig.TemplatesPath) == "" {
		if sumeruAbs != "" {
			AppConfig.TemplatesPath = filepath.Join(sumeruAbs, segCore, segEngine, segTemplates)
		} else {
			AppConfig.TemplatesPath = relPathDefaultTemplates
		}
	}

	userRoots, err := splitAbsAddonPathSegments(AppConfig.AddonsPath, ConfigFileDir)
	if err != nil {
		return err
	}
	AppConfig.AddonPaths = userRoots

	return resolveConfigPathFields(sumeruAbs)
}

func resolveConfigPathFields(sumeruRepo string) error {
	var err error
	for _, ptr := range []*string{&AppConfig.AssetsPath, &AppConfig.TemplatesPath, &AppConfig.BrandCSS, &AppConfig.LogoPath, &AppConfig.LogFile} {
		if strings.TrimSpace(*ptr) == "" {
			continue
		}
		p := strings.TrimSpace(*ptr)
		var abs string
		if filepath.IsAbs(p) {
			abs = filepath.Clean(p)
		} else if sumeruRepo != "" && strings.HasPrefix(filepath.ToSlash(filepath.Clean(p)), "core/") {
			// Repo-relative segment (e.g. core/engine/templates) lives under sumeru_home, not INI dir.
			if abs, err = filepath.Abs(filepath.Join(sumeruRepo, filepath.Clean(p))); err != nil {
				return err
			}
		} else {
			if abs, err = absFromConfigDir(p); err != nil {
				return err
			}
		}
		*ptr = abs
	}
	return nil
}

func absFromConfigDir(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", nil
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	if ConfigFileDir == "" {
		return filepath.Abs(p)
	}
	return filepath.Abs(filepath.Join(ConfigFileDir, p))
}

func splitAbsAddonPathSegments(raw, confDir string) ([]string, error) {
	var out []string
	for _, p := range strings.Split(raw, addonsPathDelimiter) {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var ap string
		var err error
		if filepath.IsAbs(p) {
			ap = filepath.Clean(p)
		} else {
			if confDir == "" {
				confDir, err = os.Getwd()
				if err != nil {
					return nil, err
				}
			}
			if ap, err = filepath.Abs(filepath.Join(confDir, p)); err != nil {
				return nil, err
			}
		}
		out = append(out, ap)
	}
	return out, nil
}
