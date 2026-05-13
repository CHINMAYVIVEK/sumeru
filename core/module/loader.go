package module

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sumeru/core/base/platformmsg"
	"sumeru/core/engine/parser"
	"sumeru/core/orm"
	"sync"
)

const coreModule = "base"

type Manifest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"` // optional; Apps / shell label
	Version     string   `json:"version"`
	Depends     []string `json:"depends"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Data        []string `json:"data"`        // XML files to load
	Application *bool    `json:"application"` // nil = true (show in Apps)
}

func (m *Manifest) IsApplication() bool {
	if m.Application == nil {
		return true
	}
	return *m.Application
}

type Addon struct {
	Manifest Manifest
	Path     string
	Menus    []parser.MenuItem
}

var (
	LoadedAddons     = map[string]*Addon{}
	DiscoveredAddons = map[string]*Addon{}
	MenuRegistry     = []parser.MenuItem{}
	installMu        sync.Mutex
)

// LoadAddonPaths discovers addons from multiple roots (comma-separated in config),
// prepends core/base (platform addons: base, user, company), syncs ir.module rows, then loads XML for installed & active modules.
// Later roots override earlier ones for the same technical module name.
func LoadAddonPaths(paths []string) error {
	var roots []string
	for _, p := range paths {
		if strings.TrimSpace(p) != "" {
			roots = append(roots, strings.TrimSpace(p))
		}
	}
	if len(roots) == 0 {
		return fmt.Errorf("no addon roots configured")
	}

	discovered, err := DiscoverAddonRoots(roots)
	if err != nil {
		return err
	}
	if err := ValidateDiscoveredAddons(discovered); err != nil {
		return err
	}
	DiscoveredAddons = discovered
	for k, v := range discovered {
		LoadedAddons[k] = v
	}

	if err := syncIrModuleRows(discovered); err != nil {
		return err
	}

	order, err := sortAddonsTopo(discovered)
	if err != nil {
		return err
	}

	for _, addon := range order {
		if !shouldSyncData(addon.Manifest.Name) {
			continue
		}
		missing, err := missingInstalledDependencies(addon.Manifest.Name)
		if err != nil {
			return fmt.Errorf("addon %q: check dependencies: %w", addon.Manifest.Name, err)
		}
		if len(missing) > 0 {
			continue
		}
		if err := addon.SyncToDB(); err != nil {
			fmt.Printf(platformmsg.FmtErrorSyncingAddon, addon.Manifest.Name, err)
		} else {
			fmt.Printf(platformmsg.FmtLoadedAddonData, addon.Manifest.Name, addon.Manifest.Version)
		}
	}

	return nil
}

// DiscoverAddonRoots scans each root for subdirectories containing manifest.json
// and returns the merged map keyed by manifest name (later roots override earlier).
func DiscoverAddonRoots(paths []string) (map[string]*Addon, error) {
	out := map[string]*Addon{}
	for _, root := range paths {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, fmt.Errorf("addons root %q: %w", root, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			addonPath := filepath.Join(root, entry.Name())
			manifestPath := filepath.Join(addonPath, "manifest.json")
			if _, err := os.Stat(manifestPath); err != nil {
				continue
			}
			manifest, err := parseManifest(manifestPath)
			if err != nil {
				return nil, fmt.Errorf("%s: manifest.json: %w", addonPath, err)
			}
			if prev, ok := out[manifest.Name]; ok {
				fmt.Printf(platformmsg.FmtAddonOverrideNotice, manifest.Name, addonPath, prev.Path)
			}
			out[manifest.Name] = &Addon{Manifest: *manifest, Path: addonPath}
		}
	}
	return out, nil
}

func countIrModules() (int, error) {
	row := orm.DB.QueryRow("SELECT COUNT(*) FROM " + orm.GetTableName("ir.module"))
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func displayNameForAddon(name string) string {
	if name == "" {
		return name
	}
	switch name {
	case "company":
		return "Companies"
	case "user":
		return "Users"
	case "base":
		return "General Settings"
	default:
		return strings.ToUpper(name[:1]) + name[1:]
	}
}

func irModuleDisplayName(a *Addon) string {
	if s := strings.TrimSpace(a.Manifest.DisplayName); s != "" {
		return s
	}
	return displayNameForAddon(a.Manifest.Name)
}

// syncIrModuleRows upserts registry metadata. New DB: all modules start installed (bootstrap).
// New addon on disk later: inserted as uninstalled until user installs from Apps.
func syncIrModuleRows(discovered map[string]*Addon) error {
	n, err := countIrModules()
	if err != nil {
		return err
	}
	bootstrap := n == 0

	menuCount := 0
	if err := orm.DB.QueryRow(`SELECT COUNT(*) FROM ` + orm.GetTableName("ir.ui.menu")).Scan(&menuCount); err != nil {
		return fmt.Errorf("count menus: %w", err)
	}
	legacyBootstrap := bootstrap && menuCount > 0

	for _, addon := range discovered {
		_, err := orm.SearchOne("ir.module", map[string]interface{}{"name": addon.Manifest.Name})
		if err == sql.ErrNoRows {
			state := "uninstalled"
			if bootstrap {
				switch {
				case addon.Manifest.Name == coreModule || !addon.Manifest.IsApplication():
					state = "installed"
				case legacyBootstrap:
					state = "installed"
				default:
					state = "uninstalled"
				}
			}
			_, err = orm.Create(orm.IrModule{}, map[string]interface{}{
				"name":         addon.Manifest.Name,
				"display_name": irModuleDisplayName(addon),
				"author":       addon.Manifest.Author,
				"version":      addon.Manifest.Version,
				"description":  addon.Manifest.Description,
				"state":        state,
				"application":  addon.Manifest.IsApplication(),
				"active":       true,
			})
			if err != nil {
				return fmt.Errorf("create ir.module %s: %w", addon.Manifest.Name, err)
			}
			continue
		}
		if err != nil {
			return err
		}
		_, err = orm.DB.Exec(
			`UPDATE `+orm.GetTableName("ir.module")+
				` SET display_name = $1, author = $2, version = $3, description = $4, application = $5 WHERE name = $6`,
			irModuleDisplayName(addon),
			addon.Manifest.Author,
			addon.Manifest.Version,
			addon.Manifest.Description,
			addon.Manifest.IsApplication(),
			addon.Manifest.Name,
		)
		if err != nil {
			return fmt.Errorf("update ir.module %s: %w", addon.Manifest.Name, err)
		}
	}
	return nil
}

func moduleRow(name string) (map[string]interface{}, error) {
	return orm.SearchOne("ir.module", map[string]interface{}{"name": name})
}

func shouldSyncData(moduleName string) bool {
	row, err := moduleRow(moduleName)
	if err != nil {
		return false
	}
	state, _ := row["state"].(string)
	active, _ := row["active"].(bool)
	if !active {
		return false
	}
	return state == "installed"
}

func sortAddonsTopo(discovered map[string]*Addon) ([]*Addon, error) {
	remaining := make(map[string]*Addon)
	for k, v := range discovered {
		remaining[k] = v
	}
	var out []*Addon
	for len(remaining) > 0 {
		var candidates []string
		for name := range remaining {
			a := remaining[name]
			satisfied := true
			for _, dep := range a.Manifest.Depends {
				dep = strings.TrimSpace(dep)
				if dep == "" || dep == name {
					continue
				}
				if _, has := discovered[dep]; !has {
					continue
				}
				if !containsAddonName(out, dep) {
					satisfied = false
					break
				}
			}
			if satisfied {
				candidates = append(candidates, name)
			}
		}
		if len(candidates) == 0 {
			keys := make([]string, 0, len(remaining))
			for k := range remaining {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			candidates = keys[:1]
		}
		sort.Strings(candidates)
		pick := candidates[0]
		out = append(out, remaining[pick])
		delete(remaining, pick)
	}
	return out, nil
}

func containsAddonName(list []*Addon, name string) bool {
	for _, a := range list {
		if a.Manifest.Name == name {
			return true
		}
	}
	return false
}

// viewArchXML persists the full parsed view (header, sheet, notebook, etc.) for ir.ui.view.arch.
func viewArchXML(v *parser.View) string {
	if v == nil {
		return "<view/>"
	}
	b, err := xml.Marshal(v)
	if err != nil {
		// Minimal fallback so DB still gets a row
		return fmt.Sprintf("<view model=\"%s\" type=\"%s\"></view>", v.Model, v.Type)
	}
	return string(b)
}

func (a *Addon) SyncToDB() error {
	for _, model := range orm.Registry {
		_, err := orm.Upsert(orm.IrModel{}, map[string]interface{}{
			"name":  model.ModelName(),
			"model": model.ModelName(),
		}, "name")
		if err != nil {
			return err
		}
	}

	modName := a.Manifest.Name
	var inheritQueue []parser.Record

	for _, xmlFile := range a.Manifest.Data {
		xmlPath := filepath.Join(a.Path, xmlFile)
		if _, err := os.Stat(xmlPath); err != nil {
			fmt.Printf(platformmsg.FmtDataFileMissing, xmlFile, modName)
			continue
		}

		viewData, err := parser.ParseViewList(xmlPath)
		if err == nil && (len(viewData.Records) > 0 || len(viewData.Views) > 0) {
			for _, rec := range viewData.Records {
				if rec.Model == "ir.actions.act_window" {
					fm := parser.RecordFieldMap(rec)
					vals := map[string]interface{}{}
					for k, v := range fm {
						vals[k] = v
					}
					if _, ok := vals["name"]; !ok || vals["name"] == "" {
						vals["name"] = rec.ID
					}
					id, err := orm.Upsert(orm.IrActionsActWindow{}, vals, "name")
					if err == nil {
						orm.Upsert(orm.IrModelData{}, map[string]interface{}{
							"module": modName,
							"name":   rec.ID,
							"model":  "ir.actions.act_window",
							"res_id": id,
						}, "name")
					}
				}
				if rec.Model == "ir.ui.view" {
					if strings.TrimSpace(parser.RecordFieldMap(rec)["inherit_id"]) != "" {
						inheritQueue = append(inheritQueue, rec)
					}
				}
				syncGenericRegistryRecord(modName, rec)
			}

			for _, v := range viewData.Views {
				arch := viewArchXML(&v)

				id, err := orm.Upsert(orm.IrUiView{}, map[string]interface{}{
					"name":  v.Model + "." + v.Type,
					"model": v.Model,
					"type":  v.Type,
					"arch":  arch,
				}, "name")
				if err == nil {
					orm.Upsert(orm.IrModelData{}, map[string]interface{}{
						"module": modName,
						"name":   v.ID,
						"model":  "ir.ui.view",
						"res_id": id,
					}, "name")
				}
			}
			continue
		}

		menus, err := parser.ParseMenus(xmlPath)
		if err == nil && len(menus) > 0 {
			for _, m := range menus {
				vals := map[string]interface{}{
					"name":     m.Name,
					"sequence": m.Sequence,
					"module":   modName,
				}

				if m.Action != "" {
					actionID, _, _ := orm.ResolveXmlId(m.Action)
					if actionID != 0 {
						vals["action_id"] = actionID
					}
				}

				if w := sanitizeWebIcon(m.WebIcon); w != "" {
					vals["web_icon"] = w
				}

				if m.ParentID != "" {
					parentID, _, _ := orm.ResolveXmlId(m.ParentID)
					if parentID != 0 {
						vals["parent_id"] = parentID
					}
				}

				id, err := orm.Upsert(orm.IrUiMenu{}, vals, "name")
				if err == nil {
					orm.Upsert(orm.IrModelData{}, map[string]interface{}{
						"module": modName,
						"name":   m.ID,
						"model":  "ir.ui.menu",
						"res_id": id,
					}, "name")
				}
			}
		}
	}

	for _, rec := range inheritQueue {
		if err := applyIrUIViewInherit(modName, rec); err != nil {
			fmt.Printf(platformmsg.FmtViewInheritWarning, modName, rec.ID, err)
		}
	}

	return nil
}

func syncGenericRegistryRecord(modName string, rec parser.Record) {
	if rec.Model == "ir.actions.act_window" || rec.Model == "ir.ui.view" {
		return
	}
	if strings.HasPrefix(rec.Model, "ir.") {
		return
	}
	inst, ok := orm.Registry[rec.Model]
	if !ok || inst == nil {
		return
	}
	fmStr := parser.RecordFieldMap(rec)
	if len(fmStr) == 0 {
		return
	}
	vals := map[string]interface{}{}
	for k, v := range fmStr {
		vals[k] = convertRecordScalar(rec.Model, k, v)
	}
	conflict := "name"
	if rec.Model == "res.users" {
		conflict = "login"
	}
	if _, ok := vals[conflict]; !ok {
		return
	}
	id, err := orm.Upsert(inst, vals, conflict)
	if err != nil {
		fmt.Printf(platformmsg.FmtGenericUpsertWarn, rec.Model, rec.ID, err)
		return
	}
	_, _ = orm.Upsert(orm.IrModelData{}, map[string]interface{}{
		"module": modName,
		"name":   rec.ID,
		"model":  rec.Model,
		"res_id": id,
	}, "name")
}

func sanitizeWebIcon(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return ""
		}
	}
	return s
}

func convertRecordScalar(model, col, raw string) interface{} {
	s := strings.TrimSpace(raw)
	if col == "active" || strings.HasSuffix(col, "_active") {
		if b, err := strconv.ParseBool(s); err == nil {
			return b
		}
		return strings.EqualFold(s, "true") || s == "1"
	}
	return s
}

func parseManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// --- Install / uninstall / activate ---

// InstallModuleByName marks the module installed and loads its XML (and dependencies first).
func InstallModuleByName(name string) error {
	installMu.Lock()
	defer installMu.Unlock()
	return installModuleUnlocked(name)
}

func installModuleUnlocked(name string) error {
	if name == "" {
		return fmt.Errorf("module name required")
	}
	a, ok := DiscoveredAddons[name]
	if !ok {
		return fmt.Errorf("unknown module %q", name)
	}

	for _, dep := range a.Manifest.Depends {
		dep = strings.TrimSpace(dep)
		if dep == "" || dep == a.Manifest.Name {
			continue
		}
		if _, has := DiscoveredAddons[dep]; !has {
			continue
		}
		row, err := moduleRow(dep)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("dependency %q is not registered", dep)
			}
			return err
		}
		if moduleStateString(row) != "installed" {
			if err := installModuleUnlocked(dep); err != nil {
				return fmt.Errorf("install dependency %q: %w", dep, err)
			}
		}
	}

	if _, err := orm.DB.Exec(
		`UPDATE `+orm.GetTableName("ir.module")+` SET state = 'installed', active = true WHERE name = $1`,
		name,
	); err != nil {
		return err
	}

	if err := a.SyncToDB(); err != nil {
		return err
	}
	return nil
}

func moduleStateString(row map[string]interface{}) string {
	if row == nil {
		return ""
	}
	switch v := row["state"].(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}

// missingInstalledDependencies lists manifest depends that are not installed in ir.module
// (or not registered). On-disk deps missing from DiscoveredAddons are ignored.
func missingInstalledDependencies(name string) ([]string, error) {
	a, ok := DiscoveredAddons[name]
	if !ok {
		return nil, nil
	}
	var missing []string
	for _, dep := range a.Manifest.Depends {
		dep = strings.TrimSpace(dep)
		if dep == "" || dep == name {
			continue
		}
		if _, has := DiscoveredAddons[dep]; !has {
			continue
		}
		row, err := moduleRow(dep)
		if err != nil {
			if err == sql.ErrNoRows {
				missing = append(missing, dep)
				continue
			}
			return nil, err
		}
		if moduleStateString(row) != "installed" {
			missing = append(missing, dep)
		}
	}
	return missing, nil
}

// UninstallModuleByName removes XML-linked metadata for the module and marks it uninstalled.
func UninstallModuleByName(name string) error {
	installMu.Lock()
	defer installMu.Unlock()

	if name == coreModule {
		return fmt.Errorf("cannot uninstall core module %q", coreModule)
	}
	if _, ok := DiscoveredAddons[name]; !ok {
		return fmt.Errorf("unknown module %q", name)
	}

	if dep, err := installedModuleDependingOn(name); err != nil {
		return err
	} else if dep != "" {
		return fmt.Errorf("module %q is required by installed module %q; uninstall that first", name, dep)
	}

	if err := deleteModuleMetadata(name); err != nil {
		return err
	}

	if _, err := orm.DB.Exec(
		`UPDATE `+orm.GetTableName("ir.module")+` SET state = 'uninstalled', active = true WHERE name = $1`,
		name,
	); err != nil {
		return err
	}
	return nil
}

func installedModuleDependingOn(target string) (string, error) {
	rows, err := orm.DB.Query(
		`SELECT name, state FROM `+orm.GetTableName("ir.module")+` WHERE state = 'installed' AND name <> $1`,
		target,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var modName, state string
		if err := rows.Scan(&modName, &state); err != nil {
			return "", err
		}
		a, ok := DiscoveredAddons[modName]
		if !ok {
			continue
		}
		for _, d := range a.Manifest.Depends {
			if strings.TrimSpace(d) == target {
				return modName, nil
			}
		}
	}
	return "", rows.Err()
}

func deleteModuleMetadata(moduleName string) error {
	models := []string{"ir.ui.menu", "ir.ui.view", "ir.actions.act_window"}
	for _, model := range models {
		tbl := orm.GetTableName(model)
		q := `DELETE FROM ` + tbl + ` WHERE id IN (SELECT res_id FROM ` + orm.GetTableName("ir.model.data") + ` WHERE module = $1 AND model = $2)`
		if _, err := orm.DB.Exec(q, moduleName, model); err != nil {
			return fmt.Errorf("delete %s: %w", model, err)
		}
	}
	if _, err := orm.DB.Exec(`DELETE FROM `+orm.GetTableName("ir.model.data")+` WHERE module = $1`, moduleName); err != nil {
		return err
	}
	return nil
}

// SetModuleActive toggles visibility of menus for an installed module without removing data.
func SetModuleActive(name string, active bool) error {
	installMu.Lock()
	defer installMu.Unlock()

	if name == coreModule && !active {
		return fmt.Errorf("cannot deactivate core module %q", coreModule)
	}
	if _, ok := DiscoveredAddons[name]; !ok {
		return fmt.Errorf("unknown module %q", name)
	}

	row, err := moduleRow(name)
	if err != nil {
		return err
	}
	if moduleStateString(row) != "installed" {
		return fmt.Errorf("module %q is not installed; activate/install it first", name)
	}

	if _, err := orm.DB.Exec(
		`UPDATE `+orm.GetTableName("ir.module")+` SET active = $1 WHERE name = $2`,
		active, name,
	); err != nil {
		return err
	}
	return nil
}

// ListModules returns ir.module rows for the Apps UI (non-application modules included for completeness).
func ListModules() ([]map[string]interface{}, error) {
	rows, err := orm.DB.Query(
		`SELECT id, name, display_name, author, version, description, state, application, active FROM ` +
			orm.GetTableName("ir.module") + ` ORDER BY application DESC, name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]interface{}
	cols := []string{"id", "name", "display_name", "author", "version", "description", "state", "application", "active"}

	for rows.Next() {
		var id int64
		var name, display, author, version, state string
		var desc sql.NullString
		var application, active bool
		if err := rows.Scan(&id, &name, &display, &author, &version, &desc, &state, &application, &active); err != nil {
			return nil, err
		}
		m := make(map[string]interface{})
		m["id"] = id
		m[cols[1]] = name
		m[cols[2]] = display
		m[cols[3]] = author
		m[cols[4]] = version
		if desc.Valid {
			m[cols[5]] = desc.String
		} else {
			m[cols[5]] = ""
		}
		m[cols[6]] = state
		m[cols[7]] = application
		m[cols[8]] = active
		out = append(out, m)
	}
	return out, rows.Err()
}
