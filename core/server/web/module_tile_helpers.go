package web

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/module"
	"sumeru/core/orm"
)

type appTile struct {
	Name         string
	DisplayName  string
	Version      string
	Description  string
	Author       string
	IconLetter   string
	IconHue      int
	OpenMenuHref string
}

func loadInstalledAppTiles(ctx context.Context, searchQ string) ([]appTile, error) {
	raw, err := module.ListModules(ctx)
	if err != nil {
		return nil, err
	}
	var tiles []appTile
	for _, row := range raw {
		name := orm.AsString(row["name"])
		if name == "" {
			continue
		}
		if !orm.AsBool(row["application"]) {
			continue
		}
		if orm.AsString(row["state"]) != "installed" {
			continue
		}
		if !orm.AsBool(row["active"]) {
			continue
		}
		dn := orm.AsString(row["display_name"])
		if dn == "" {
			dn = name
		}
		desc := orm.AsString(row["description"])
		if searchQ != "" && !homeSearchMatches(searchQ, name, dn, desc) {
			continue
		}
		openHref := "/web/apps"
		if mid := rootMenuIDForModule(ctx, name); mid > 0 {
			openHref = fmt.Sprintf("/web?menu_id=%d", mid)
		}
		tiles = append(tiles, appTile{
			Name:         name,
			DisplayName:  dn,
			Version:      orm.AsString(row["version"]),
			Description:  desc,
			Author:       orm.AsString(row["author"]),
			IconLetter:   iconLetterFromName(dn),
			IconHue:      iconHueFromString(name),
			OpenMenuHref: openHref,
		})
	}
	return tiles, nil
}

func iconLetterFromName(name string) string {
	if r := []rune(strings.TrimSpace(name)); len(r) > 0 {
		return strings.ToUpper(string(r[0]))
	}
	return "?"
}

func rootMenuIDForModule(ctx context.Context, moduleName string) int {
	if orm.DB == nil || strings.TrimSpace(moduleName) == "" {
		return 0
	}
	tbl := orm.GetTableName("sys.menu")
	q := `SELECT id FROM ` + tbl + ` WHERE module = $1 AND parent_id IS NULL ORDER BY sequence ASC, id ASC LIMIT 1`
	var id int
	if err := orm.DB.QueryRowContext(ctx, q, strings.TrimSpace(moduleName)).Scan(&id); err != nil {
		return 0
	}
	return id
}

func iconHueFromString(s string) int {
	h := 265
	for _, c := range strings.TrimSpace(s) {
		h = (h*31 + int(c)) % 360
		if h < 0 {
			h += 360
		}
	}
	return h
}
