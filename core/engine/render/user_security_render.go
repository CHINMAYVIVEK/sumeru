package render

import (
	"context"
	"fmt"
	"html/template"
	"sort"
	"strings"

	"sumeru/core/orm"
)

func writeResUsersSecuritySection(ctx context.Context, sb *strings.Builder, vr *ViewRecordData, ro bool) {
	uid := orm.SecurityUID(ctx)
	if uid <= 0 {
		return
	}
	if err := orm.CheckModelAccess(ctx, uid, "core.group", "read"); err != nil {
		return
	}
	isAdmin := orm.UserHasGroupXML(ctx, uid, "base.group_system")
	// Only system admins may edit access rights; others see assigned groups read-only.
	editable := isAdmin && !ro

	sb.WriteString(`<div class="sum-security-card">`)
	sb.WriteString(`<h3 class="sum-security-title">Access rights</h3>`)
	if editable {
		sb.WriteString(`<p class="sum-security-muted">Only installed applications are listed. Manager includes User rights for that app. System admin includes all module Managers. Portal and Public are mutually exclusive with Internal User.</p>`)
		sb.WriteString(`<input type="hidden" name="security_groups_touched" value="1"/>`)
		sb.WriteString(`<div data-password-match class="sum-field-widget sum-field-widget--spaced">`)
		sb.WriteString(`<label class="sum-field-label" for="password_plain">New password</label>`)
		sb.WriteString(`<input class="sum-field-input" id="password_plain" name="password_plain" type="password" autocomplete="new-password" placeholder="Leave blank to keep current" data-password-primary />`)
		sb.WriteString(`<label class="sum-field-label" for="password_plain_confirm">Confirm new password</label>`)
		sb.WriteString(`<input class="sum-field-input" id="password_plain_confirm" name="password_plain_confirm" type="password" autocomplete="new-password" placeholder="Repeat new password" data-password-confirm />`)
		sb.WriteString(`<p class="sum-field-hint" data-password-match-hint role="alert" aria-live="polite" hidden></p>`)
		sb.WriteString(`</div>`)
	} else {
		sb.WriteString(`<p class="sum-security-muted">Assigned access for this user. Only an administrator can change access rights.</p>`)
	}

	selected := map[int]struct{}{}
	if vr.RecordID > 0 {
		rel := orm.MustQuotedTableName("core.group.user.rel")
		rows, err := orm.DB.QueryContext(ctx, `SELECT group_id FROM `+rel+` WHERE user_id = $1`, vr.RecordID)
		if err == nil {
			for rows.Next() {
				var gid int
				if err := rows.Scan(&gid); err == nil {
					selected[gid] = struct{}{}
				}
			}
			rows.Close()
		}
	}

	groups, err := orm.ListAllGroupRows(ctx)
	if err != nil || len(groups) == 0 {
		sb.WriteString(`<p class="sum-security-muted">No groups defined.</p></div>`)
		return
	}

	typeXML := map[string]int{}
	for _, x := range []string{"base.group_user", "base.group_portal", "base.group_public", "base.group_system"} {
		if gid, _, err := orm.ResolveXmlId(ctx, x); err == nil && gid > 0 {
			typeXML[x] = gid
		}
	}

	if !editable {
		var names []string
		for _, g := range groups {
			id, _ := orm.CoerceInt64(g["id"])
			if _, ok := selected[int(id)]; ok {
				names = append(names, orm.AsString(g["name"]))
			}
		}
		sort.Strings(names)
		if len(names) == 0 {
			sb.WriteString(`<p class="sum-security-body">No access groups assigned.</p></div>`)
			return
		}
		sb.WriteString(`<ul class="sum-security-assigned-list">`)
		for _, nm := range names {
			sb.WriteString(`<li>` + template.HTMLEscapeString(nm) + `</li>`)
		}
		sb.WriteString(`</ul></div>`)
		return
	}

	installed, _ := orm.InstalledModuleNames(ctx)
	if installed == nil {
		installed = map[string]struct{}{}
	}
	// Kernel groups always visible for admin assignment.
	installed["base"] = struct{}{}
	groupModule := loadCoreGroupModules(ctx)

	// User type radios
	sb.WriteString(`<div class="sum-security-section">`)
	sb.WriteString(`<h4 class="sum-security-section-title">User type</h4>`)
	sb.WriteString(`<div class="sum-security-check-list">`)
	typeChoices := []struct{ xml, label string }{
		{"base.group_user", "Internal User"},
		{"base.group_portal", "Portal"},
		{"base.group_public", "Public"},
	}
	chosenType := ""
	for _, c := range typeChoices {
		if gid, ok := typeXML[c.xml]; ok {
			if _, sel := selected[gid]; sel {
				chosenType = c.xml
				break
			}
		}
	}
	if chosenType == "" {
		chosenType = "base.group_user"
	}
	for _, c := range typeChoices {
		gid := typeXML[c.xml]
		if gid == 0 {
			continue
		}
		chk := ""
		if c.xml == chosenType {
			chk = ` checked`
		}
		sb.WriteString(`<label class="sum-security-check-row">`)
		sb.WriteString(fmt.Sprintf(`<input type="radio" name="security_user_type" value="%d"%s /> `, gid, chk))
		sb.WriteString(template.HTMLEscapeString(c.label))
		sb.WriteString(`</label>`)
	}
	sb.WriteString(`</div></div>`)

	type gRow struct {
		ID       int
		Name     string
		CatID    int
		CatName  string
		Sequence int
	}
	var rows []gRow
	catNames := map[int]string{}
	if cats, err := orm.Search(ctx, "sys.module.category", nil); err == nil {
		for _, c := range cats {
			cid, _ := orm.CoerceInt64(c["id"])
			catNames[int(cid)] = orm.AsString(c["name"])
		}
	}
	skipTypes := map[int]struct{}{}
	for _, x := range []string{"base.group_user", "base.group_portal", "base.group_public"} {
		if gid, ok := typeXML[x]; ok {
			skipTypes[gid] = struct{}{}
		}
	}
	for _, g := range groups {
		gid, ok := orm.CoerceInt64(g["id"])
		if !ok {
			continue
		}
		if _, skip := skipTypes[int(gid)]; skip {
			continue
		}
		mod := strings.TrimSpace(groupModule[int(gid)])
		if mod == "" {
			mod = "base"
		}
		if _, ok := installed[mod]; !ok {
			continue
		}
		cid, _ := orm.CoerceInt64(g["category_id"])
		seq, _ := orm.CoerceInt64(g["sequence"])
		cn := catNames[int(cid)]
		if cn == "" {
			cn = "Other"
		}
		rows = append(rows, gRow{ID: int(gid), Name: orm.AsString(g["name"]), CatID: int(cid), CatName: cn, Sequence: int(seq)})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CatName != rows[j].CatName {
			return rows[i].CatName < rows[j].CatName
		}
		if rows[i].Sequence != rows[j].Sequence {
			return rows[i].Sequence < rows[j].Sequence
		}
		return rows[i].Name < rows[j].Name
	})

	curCat := ""
	for _, g := range rows {
		if g.CatName != curCat {
			if curCat != "" {
				sb.WriteString(`</div></div>`)
			}
			curCat = g.CatName
			sb.WriteString(`<div class="sum-security-section">`)
			sb.WriteString(`<h4 class="sum-security-section-title">` + template.HTMLEscapeString(curCat) + `</h4>`)
			sb.WriteString(`<div class="sum-security-check-list">`)
		}
		_, checked := selected[g.ID]
		chk := ""
		if checked {
			chk = ` checked`
		}
		sb.WriteString(`<label class="sum-security-check-row">`)
		sb.WriteString(fmt.Sprintf(`<input type="checkbox" name="security_group_ids" value="%d"%s />`, g.ID, chk))
		sb.WriteString(template.HTMLEscapeString(g.Name))
		sb.WriteString(`</label>`)
	}
	if curCat != "" {
		sb.WriteString(`</div></div>`)
	}
	sb.WriteString(`</div>`)
}

// loadCoreGroupModules maps core.group id → declaring module from sys.model.data.
func loadCoreGroupModules(ctx context.Context) map[int]string {
	out := map[int]string{}
	if orm.DB == nil {
		return out
	}
	tbl := orm.MustQuotedTableName("sys.model.data")
	rows, err := orm.DB.QueryContext(ctx,
		`SELECT core_id, module FROM `+tbl+` WHERE model = $1 AND core_id IS NOT NULL AND core_id > 0`,
		"core.group")
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var mod string
		if err := rows.Scan(&id, &mod); err != nil {
			continue
		}
		mod = strings.TrimSpace(mod)
		if id > 0 && mod != "" {
			out[id] = mod
		}
	}
	return out
}
