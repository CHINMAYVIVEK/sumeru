package web

import (
	"context"
	"fmt"

	"sumeru/core/engine/render"
	"sumeru/core/module"
	"sumeru/core/orm"
)

func loadInstalledAppTiles(ctx context.Context, searchQ string) ([]render.AppTile, error) {
	raw, err := module.ListModules(ctx)
	if err != nil {
		return nil, err
	}
	var tiles []render.AppTile
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
		displayName := orm.AsString(row["display_name"])
		if displayName == "" {
			displayName = name
		}
		description := orm.AsString(row["description"])
		if searchQ != "" && !homeSearchMatches(searchQ, name, displayName, description) {
			continue
		}
		openHref := "/web/apps"
		if menuID := render.RootMenuIDForModule(ctx, name); menuID > 0 {
			openHref = fmt.Sprintf("/web?menu_id=%d", menuID)
		}
		tiles = append(tiles, render.AppTile{
			Name:         name,
			DisplayName:  displayName,
			Version:      orm.AsString(row["version"]),
			Description:  description,
			Author:       orm.AsString(row["author"]),
			IconLetter:   render.IconLetterFromName(displayName),
			IconHue:      render.IconHueFromString(name),
			OpenMenuHref: openHref,
		})
	}
	return tiles, nil
}
