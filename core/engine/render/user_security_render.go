package render

import (
	"context"
	"fmt"
	"html/template"
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
	sb.WriteString(`<div class="sum-security-card">`)
	sb.WriteString(`<h3 class="sum-security-title">Access rights</h3>`)
	if !ro {
		sb.WriteString(`<input type="hidden" name="security_groups_touched" value="1"/>`)
		sb.WriteString(`<div class="sum-field-widget sum-field-widget--spaced">`)
		sb.WriteString(`<label class="sum-field-label" for="password_plain">New password</label>`)
		sb.WriteString(`<input class="sum-field-input" id="password_plain" name="password_plain" type="password" autocomplete="new-password" placeholder="Leave blank to keep current" />`)
		sb.WriteString(`</div>`)
	}
	selected := map[int]struct{}{}
	if vr.RecordID > 0 {
		rel := orm.GetTableName("core.group.user.rel")
		rows, err := orm.DB.Query(`SELECT group_id FROM `+rel+` WHERE user_id = $1`, vr.RecordID)
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
	if ro {
		var names []string
		for _, g := range groups {
			id, _ := orm.CoerceInt64(g["id"])
			if _, ok := selected[int(id)]; ok {
				names = append(names, orm.AsString(g["name"]))
			}
		}
		sb.WriteString(`<p class="sum-security-body">` + template.HTMLEscapeString(strings.Join(names, ", ")) + `</p></div>`)
		return
	}
	sb.WriteString(`<div class="sum-security-check-list">`)
	for _, g := range groups {
		gid, ok := orm.CoerceInt64(g["id"])
		if !ok {
			continue
		}
		nm := orm.AsString(g["name"])
		_, checked := selected[int(gid)]
		chk := ""
		if checked {
			chk = ` checked`
		}
		sb.WriteString(`<label class="sum-security-check-row">`)
		sb.WriteString(fmt.Sprintf(`<input type="checkbox" name="security_group_ids" value="%d"%s />`, int(gid), chk))
		sb.WriteString(template.HTMLEscapeString(nm))
		sb.WriteString(`</label>`)
	}
	sb.WriteString(`</div></div>`)
}
