package render

import (
	"fmt"
	"html/template"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"sumeru/core/engine/parser"
	"sumeru/core/mail"
	"sumeru/core/orm"
)

type PageData struct {
	Title                  string // legacy / diagnostics; prefer ViewBreadcrumb for UI
	ViewBreadcrumb         string // human label for breadcrumb (not the technical model id)
	AppName                string // product display name (browser tab suffix, header)
	ModuleName             string
	Content                template.HTML
	TopMenus               []parser.MenuItem
	SidebarMenus           []SidebarMenu
	ActiveModuleID         string
	ActiveMenuID           string
	ViewStylesheetURLs     []string
	AppsNavActive          bool
	ExtraStylesheetURLs    []string
	LogoURL                string
	ShellCompany           string
	ShellUser              string
	UserInitial            string          // first letter for avatar
	ViewTabs               []ViewSwitchTab // workspace view switcher; empty hides toolbar
	HideBreadcrumbViewTabs bool            // true on tree/list (tabs appear in list control panel)

	// Right activity panel: Log tab (audit); Messages tab HTML set in RenderView when chatter applies.
	ActivityEnabled         bool
	ActivityLogItems        []ActivityItem
	ActivityContextModel    string
	ActivityContextRecordID int64
	ActivityPanelChatter    bool
	ActivityChatterHTML     template.HTML
}

// ActivityItem is one line in the shell activity feed.
type ActivityItem struct {
	Meta string // author · relative time
	Body string
}

type SidebarMenu struct {
	ID       string
	Name     string
	Sequence int
	SubMenus []parser.MenuItem
}

// ViewRecordData carries rows loaded from the ORM for HTML rendering.
type ViewRecordData struct {
	ActionID int
	Record   map[string]interface{}
	ListRows []map[string]interface{}
	ViewTabs []ViewSwitchTab // optional; copied onto PageData for base layout

	// Workspace form chrome (/web): Edit / Save / Cancel and POST save target.
	ResModel       string // e.g. res.company
	RecordID       int    // 0 = create form
	FormEditing    bool   // true when URL contains edit=1
	FormBaseQuery  string // query string for /web without leading "?" and without edit= (action, menu_id, view_type, id)
	FormSaveAction string // POST URL; default "/web/record/save"
}

func recStr(rec map[string]interface{}, name string) string {
	if rec == nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(rec[name]))
}

func isTruthyDB(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	case int32:
		return t != 0
	case int:
		return t != 0
	case float64:
		return t != 0
	case []byte:
		s := strings.ToLower(strings.TrimSpace(string(t)))
		return s == "t" || s == "true" || s == "1"
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		return s == "t" || s == "true" || s == "1"
	default:
		return false
	}
}

func rowOpenURL(actionID int, menuID string, rowID int64) string {
	q := url.Values{}
	if actionID > 0 {
		q.Set("action", fmt.Sprintf("%d", actionID))
	}
	if strings.TrimSpace(menuID) != "" {
		q.Set("menu_id", strings.TrimSpace(menuID))
	}
	q.Set("view_type", "form")
	q.Set("id", fmt.Sprintf("%d", rowID))
	return "/web?" + q.Encode()
}

func formFieldReadonly(vr *ViewRecordData) bool {
	if vr == nil || strings.TrimSpace(vr.ResModel) == "" {
		return true
	}
	if vr.RecordID == 0 {
		return false
	}
	return !vr.FormEditing
}

func workspaceFormChrome(vr *ViewRecordData) bool {
	return vr != nil && strings.TrimSpace(vr.ResModel) != ""
}

func renderWorkspaceFormChrome(sb *strings.Builder, vr *ViewRecordData) {
	if vr == nil || strings.TrimSpace(vr.ResModel) == "" {
		return
	}
	base := strings.TrimSpace(vr.FormBaseQuery)
	hasID := vr.RecordID > 0
	editQS := "edit=1"
	if base != "" {
		editQS = base + "&edit=1"
	}
	editURL := "/web?" + editQS
	cancelURL := "/web"
	if base != "" {
		if !hasID {
			if qv, err := url.ParseQuery(base); err == nil {
				qv.Set("view_type", "tree")
				qv.Del("id")
				cancelURL = "/web?" + qv.Encode()
			} else {
				cancelURL = "/web?" + base
			}
		} else {
			cancelURL = "/web?" + base
		}
	}
	saveAct := strings.TrimSpace(vr.FormSaveAction)
	if saveAct == "" {
		saveAct = "/web/record/save"
	}
	nextEsc := template.HTMLEscapeString(cancelURL)

	sb.WriteString(`<div class="sum-ws-form-chrome">`)
	if hasID && !vr.FormEditing {
		sb.WriteString(`<a href="` + template.HTMLEscapeString(editURL) + `" class="sum-ws-btn sum-ws-btn--ghost">Edit</a>`)
	}
	if hasID && vr.FormEditing {
		sb.WriteString(`<button type="submit" form="sum-workspace-record-form" class="sum-ws-btn sum-ws-btn--primary">Save</button>`)
		sb.WriteString(`<a href="` + template.HTMLEscapeString(cancelURL) + `" class="sum-ws-btn sum-ws-btn--ghost">Cancel</a>`)
	}
	if !hasID {
		sb.WriteString(`<button type="submit" form="sum-workspace-record-form" class="sum-ws-btn sum-ws-btn--primary">Save</button>`)
		sb.WriteString(`<a href="` + template.HTMLEscapeString(cancelURL) + `" class="sum-ws-btn sum-ws-btn--ghost">Cancel</a>`)
	}
	sb.WriteString(`</div>`)

	sb.WriteString(`<form id="sum-workspace-record-form" method="post" action="` + template.HTMLEscapeString(saveAct) + `" class="contents">`)
	sb.WriteString(`<input type="hidden" name="model" value="` + template.HTMLEscapeString(vr.ResModel) + `" />`)
	if hasID {
		sb.WriteString(`<input type="hidden" name="id" value="` + template.HTMLEscapeString(fmt.Sprintf("%d", vr.RecordID)) + `" />`)
	}
	sb.WriteString(`<input type="hidden" name="next" value="` + nextEsc + `" />`)
}

func renderWorkspaceFormChromeClose(sb *strings.Builder, vr *ViewRecordData) {
	if workspaceFormChrome(vr) {
		sb.WriteString(`</form>`)
	}
}

func RenderView(view *parser.View, activeMenuID, templatesDir string, recData *ViewRecordData) string {
	if recData == nil {
		recData = &ViewRecordData{}
	}
	var content string
	switch view.Type {
	case "form":
		content = RenderForm(view, recData)
	case "tree", "list":
		content = RenderTree(view, recData.ListRows, recData.ActionID, activeMenuID, recData.ViewTabs)
	case "kanban":
		content = RenderKanban(view, recData.ListRows, recData.ActionID, activeMenuID)
	case "pivot":
		content = RenderPivot(view)
	default:
		content = RenderForm(view, recData)
	}

	topMenus, sidebarMenus, activeModuleID := LoadShellMenus(activeMenuID)
	moduleName := ModuleNameForTopMenu(topMenus, activeModuleID)
	viewBC := HumanViewBreadcrumb(view.Model, view.Type)

	actCtxModel := ""
	var actCtxID int64
	if strings.EqualFold(view.Type, "form") && recData.RecordID > 0 {
		actCtxModel = strings.TrimSpace(view.Model)
		actCtxID = int64(recData.RecordID)
	}

	pageData := PageData{
		Title:                   fmt.Sprintf("%s · %s", view.Model, view.Type),
		ViewBreadcrumb:          viewBC,
		ModuleName:              moduleName,
		Content:                 template.HTML(content),
		TopMenus:                topMenus,
		SidebarMenus:            sidebarMenus,
		ActiveModuleID:          activeModuleID,
		ActiveMenuID:            activeMenuID,
		ViewStylesheetURLs:      []string{"/static/css/view-web.css"},
		ExtraStylesheetURLs:     ExtraStylesheetURLs,
		ViewTabs:                recData.ViewTabs,
		ActivityContextModel:    actCtxModel,
		ActivityContextRecordID: actCtxID,
		HideBreadcrumbViewTabs:  strings.EqualFold(view.Type, "tree") || strings.EqualFold(view.Type, "list"),
	}
	if strings.EqualFold(view.Type, "form") && view.Chatter != nil && recData.RecordID > 0 &&
		mail.CompanyChatterEnabled() && mail.CompanyActivityPanelEnabled() {
		pageData.ActivityPanelChatter = true
		var ch strings.Builder
		writeActivityChatterPanel(&ch, view.Chatter, recData, view.Model)
		pageData.ActivityChatterHTML = template.HTML(ch.String())
	}

	log.Printf("Rendering view for model %s (ActiveMenu: %s, ActiveModule: %s)", view.Model, activeMenuID, activeModuleID)

	out, err := RenderPage(templatesDir, pageData)
	if err != nil {
		log.Printf("Error rendering page: %v", err)
		return content
	}
	return out
}

func RenderForm(view *parser.View, vr *ViewRecordData) string {
	if vr == nil {
		vr = &ViewRecordData{}
	}
	record := vr.Record
	if record == nil {
		record = map[string]interface{}{}
	}
	ro := formFieldReadonly(vr)
	chrome := workspaceFormChrome(vr)

	var sb strings.Builder

	formViewClass := "o_form_view sum-form-view flex h-full min-h-0 overflow-hidden"
	if ro && vr.RecordID > 0 {
		formViewClass += " sum-form-view--readonly"
	}
	sb.WriteString(`<div class="` + formViewClass + `">`)

	sb.WriteString(`<div class="o_form_sheet_bg sum-form-sheet-bg flex-1 overflow-y-auto p-4 md:p-8">`)

	if chrome {
		renderWorkspaceFormChrome(&sb, vr)
	}

	if view.Header != nil {
		renderHeader(&sb, view.Header, record)
	}

	if view.Sheet != nil {
		renderSheet(&sb, view.Sheet, record, ro)
	} else {
		sb.WriteString(`<div class="o_form_sheet sum-form-sheet sum-form-sheet--solo">`)
		for _, f := range view.Field {
			renderField(&sb, f, record, ro)
		}
		for _, g := range view.Group {
			renderGroup(&sb, g, record, ro)
		}
		sb.WriteString(`</div>`)
	}

	if strings.TrimSpace(vr.ResModel) == "res.users" {
		writeResUsersSecuritySection(&sb, vr, ro)
	}

	if view.Footer != nil {
		renderFormFooter(&sb, view.Footer)
	}

	if chrome {
		renderWorkspaceFormChromeClose(&sb, vr)
	}

	sb.WriteString(`</div>`)

	sb.WriteString(`</div>`)

	return sb.String()
}

func renderHeader(sb *strings.Builder, h *parser.Header, record map[string]interface{}) {
	sb.WriteString(`<div class="o_form_statusbar sum-form-statusbar flex items-center justify-between p-2 mb-0 border-b-0">`)

	// Buttons
	sb.WriteString(`<div class="o_statusbar_buttons flex space-x-2">`)
	for _, b := range h.Button {
		class := "px-3 py-1.5 rounded text-sm font-bold shadow-sm transition-all "
		if b.Class == "oe_highlight" {
			class += "bg-indigo-600 text-white hover:bg-indigo-700"
		} else {
			class += "bg-white text-slate-700 border border-slate-300 hover:bg-slate-50"
		}
		sb.WriteString(fmt.Sprintf(`<button type="button" disabled class="%s opacity-60 cursor-not-allowed">%s</button>`, class, template.HTMLEscapeString(b.String)))
	}
	sb.WriteString(`</div>`)

	// Statusbar fields: show stored values only (no mock workflow states).
	sb.WriteString(`<div class="o_statusbar_status flex flex-wrap justify-end gap-2 items-center">`)
	for _, hf := range h.Field {
		val := recStr(record, hf.Name)
		if val == "" {
			continue
		}
		sb.WriteString(`<span class="px-3 py-1.5 text-xs font-semibold rounded-sm bg-slate-50 text-slate-700 border border-slate-200">`)
		sb.WriteString(template.HTMLEscapeString(val))
		sb.WriteString(`</span>`)
	}
	sb.WriteString(`</div>`)

	sb.WriteString(`</div>`)
}

func renderSheet(sb *strings.Builder, s *parser.Sheet, record map[string]interface{}, ro bool) {
	sb.WriteString(`<div class="o_form_sheet sum-form-sheet">`)

	// Render Stat Buttons if any (Usually inside a div with class oe_button_box)
	for _, d := range s.Div {
		if strings.Contains(d.Class, "oe_button_box") {
			renderButtonBox(sb, d)
		}
	}

	// Render Title/Avatar area
	for _, d := range s.Div {
		if strings.Contains(d.Class, "oe_title") {
			renderTitle(sb, d, record, ro)
		}
	}

	// Render direct fields/groups in sheet
	if ro {
		sb.WriteString(`<div class="sum-read-fields sum-read-fields--sheet mb-8">`)
	} else {
		sb.WriteString(`<div class="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">`)
	}
	for _, sep := range s.Separator {
		renderSeparator(sb, sep)
	}
	for _, lab := range s.Label {
		renderLabel(sb, lab)
	}
	for _, g := range s.Group {
		renderGroup(sb, g, record, ro)
	}
	for _, f := range s.Field {
		renderField(sb, f, record, ro)
	}
	sb.WriteString(`</div>`)

	// Render Notebooks (Tabs)
	for _, nb := range s.Notebook {
		renderNotebook(sb, nb, record, ro)
	}

	sb.WriteString(`</div>`)
}

func renderButtonBox(sb *strings.Builder, d parser.Div) {
	_ = d
	sb.WriteString(`<div class="flex justify-end mb-8 -mr-8 -mt-8 border-b border-slate-100 min-h-[1px]"></div>`)
}

func renderTitle(sb *strings.Builder, d parser.Div, record map[string]interface{}, ro bool) {
	sb.WriteString(`<div class="sum-form-title-row flex items-start gap-8 mb-10">`)

	sb.WriteString(`<div class="sum-form-avatar shrink-0">`)
	sb.WriteString(`<div class="sum-form-avatar-box w-36 h-36 bg-slate-50 border border-dashed border-slate-200 rounded-lg flex flex-col items-center justify-center text-center p-3">`)
	sb.WriteString(`<svg class="w-10 h-10 text-slate-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>`)
	sb.WriteString(`<span class="text-xs text-slate-500 leading-tight">Your logo</span>`)
	sb.WriteString(`</div></div>`)

	sb.WriteString(`<div class="flex-1 min-w-0">`)
	if ro {
		for _, h1 := range d.H1 {
			for _, f := range h1.Field {
				ph := strings.TrimSpace(f.Label)
				if ph == "" {
					ph = "Title"
				}
				v := recStr(record, f.Name)
				sb.WriteString(`<div class="text-xs font-medium text-slate-500 mb-1">` + template.HTMLEscapeString(ph) + `</div>`)
				sb.WriteString(`<div class="sum-read-hero-title text-3xl sm:text-4xl font-semibold text-slate-900 tracking-tight">` + template.HTMLEscapeString(v) + `</div>`)
			}
		}
		sb.WriteString(`<div class="mt-5 space-y-3">`)
		for _, f := range d.Field {
			icon := "M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
			if strings.Contains(f.Name, "phone") {
				icon = "M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
			}
			if strings.Contains(f.Name, "tag") {
				icon = "M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
			}
			v := recStr(record, f.Name)
			sb.WriteString(`<div class="sum-read-inline flex items-center gap-2 text-slate-700">`)
			sb.WriteString(fmt.Sprintf(`<svg class="w-4 h-4 text-slate-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="%s"></path></svg>`, icon))
			sb.WriteString(`<span class="text-sm">` + template.HTMLEscapeString(v) + `</span>`)
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
	} else {
		for _, h1 := range d.H1 {
			for _, f := range h1.Field {
				ph := template.HTMLEscapeString(f.Label)
				v := recStr(record, f.Name)
				sb.WriteString(fmt.Sprintf(`<input class="text-4xl font-bold text-slate-800 border-b border-transparent hover:border-slate-200 focus:border-indigo-500 focus:outline-none w-full bg-transparent pb-1 mb-4" placeholder="%s" name="%s" value="%s" />`,
					ph, template.HTMLEscapeString(f.Name), template.HTMLEscapeString(v)))
			}
		}
		sb.WriteString(`<div class="space-y-2">`)
		for _, f := range d.Field {
			icon := "M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
			if strings.Contains(f.Name, "phone") {
				icon = "M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
			}
			if strings.Contains(f.Name, "tag") {
				icon = "M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
			}
			v := recStr(record, f.Name)
			sb.WriteString(`<div class="flex items-center text-slate-500 group">`)
			sb.WriteString(fmt.Sprintf(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="%s"></path></svg>`, icon))
			sb.WriteString(fmt.Sprintf(`<input class="text-sm border-b border-transparent hover:border-slate-200 focus:border-indigo-500 focus:outline-none bg-transparent py-0.5" placeholder="%s" name="%s" value="%s" />`,
				template.HTMLEscapeString(f.Label), template.HTMLEscapeString(f.Name), template.HTMLEscapeString(v)))
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
	}

	sb.WriteString(`</div>`)
	sb.WriteString(`</div>`)
}

func renderNotebook(sb *strings.Builder, nb parser.Notebook, record map[string]interface{}, ro bool) {
	sb.WriteString(`<div class="o_notebook sum-notebook mt-10">`)

	sb.WriteString(`<div class="sum-notebook-tabs flex border-b border-slate-200 mb-6 gap-1" role="tablist">`)
	for i, p := range nb.Page {
		activeClass := "sum-notebook-tab"
		if i == 0 {
			activeClass += " sum-notebook-tab--active"
		}
		sb.WriteString(fmt.Sprintf(`<button type="button" role="tab" class="%s">%s</button>`, activeClass, template.HTMLEscapeString(p.Title)))
	}
	sb.WriteString(`</div>`)

	// Tabs Content
	sb.WriteString(`<div class="o_notebook_content">`)
	for i, p := range nb.Page {
		display := "none"
		if i == 0 {
			display = "block"
		}
		sb.WriteString(fmt.Sprintf(`<div class="o_notebook_page" style="display: %s">`, display))
		sb.WriteString(`<div class="grid grid-cols-1 md:grid-cols-2 gap-8">`)
		for _, sep := range p.Separator {
			renderSeparator(sb, sep)
		}
		for _, lab := range p.Label {
			renderLabel(sb, lab)
		}
		for _, g := range p.Group {
			renderGroup(sb, g, record, ro)
		}
		for _, f := range p.Field {
			renderField(sb, f, record, ro)
		}
		sb.WriteString(`</div>`)
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div>`)

	sb.WriteString(`</div>`)
}

func renderSeparator(sb *strings.Builder, sep parser.Separator) {
	t := strings.TrimSpace(sep.String)
	if t != "" {
		sb.WriteString(`<div class="o_horizontal_separator col-span-2 text-sm font-semibold text-slate-500 border-b border-slate-200 pb-2 mb-4 mt-2">` + template.HTMLEscapeString(t) + `</div>`)
	} else {
		sb.WriteString(`<hr class="col-span-2 border-slate-200 my-4"/>`)
	}
}

func renderLabel(sb *strings.Builder, lab parser.Label) {
	s := strings.TrimSpace(lab.String)
	if s == "" {
		return
	}
	sb.WriteString(`<div class="o_label col-span-2 mb-2">`)
	sb.WriteString(`<span class="text-xs font-bold text-slate-500 uppercase tracking-wide"`)
	if id := strings.TrimSpace(lab.For); id != "" {
		sb.WriteString(` for="` + template.HTMLEscapeString(id) + `"`)
	}
	sb.WriteString(`>` + template.HTMLEscapeString(s) + `</span></div>`)
}

func renderFormFooter(sb *strings.Builder, ft *parser.Footer) {
	if ft == nil || len(ft.Button) == 0 {
		return
	}
	sb.WriteString(`<div class="o_form_footer flex flex-wrap justify-end gap-2 px-8 py-4 bg-slate-50 border-t border-slate-200 rounded-b-sm">`)
	for _, b := range ft.Button {
		cls := "px-4 py-2 text-sm font-semibold rounded-sm border border-slate-300 bg-white text-slate-700 hover:bg-slate-100"
		if strings.Contains(b.Class, "oe_highlight") || strings.Contains(b.Class, "btn-primary") {
			cls = "px-4 py-2 text-sm font-semibold rounded-sm bg-indigo-600 text-white hover:bg-indigo-700"
		}
		label := template.HTMLEscapeString(b.String)
		if label == "" {
			label = template.HTMLEscapeString(b.Name)
		}
		sb.WriteString(fmt.Sprintf(`<button type="button" disabled name="%s" class="%s opacity-60 cursor-not-allowed">%s</button>`,
			template.HTMLEscapeString(b.Name), cls, label))
	}
	sb.WriteString(`</div>`)
}

// writeActivityChatterPanel renders thread + composer for the right activity panel (comments only).
func writeActivityChatterPanel(sb *strings.Builder, c *parser.Chatter, vr *ViewRecordData, viewModel string) {
	_ = c
	if !mail.CompanyChatterEnabled() || vr == nil || vr.RecordID <= 0 {
		return
	}
	model := strings.TrimSpace(viewModel)
	if model == "" {
		model = strings.TrimSpace(vr.ResModel)
	}
	if model == "" {
		return
	}

	msgs, err := mail.ListCommentsForRecord(model, int64(vr.RecordID), 120)
	if err != nil {
		log.Printf("activity chatter list %s %d: %v", model, vr.RecordID, err)
		msgs = nil
	}

	nextURL := "/web"
	if q := strings.TrimSpace(vr.FormBaseQuery); q != "" {
		nextURL = "/web?" + q
	}

	sb.WriteString(`<div class="sum-msg-shell">`)
	sb.WriteString(`<div class="sum-msg-thread">`)
	for _, m := range msgs {
		meta := strings.TrimSpace(m.Author)
		if meta == "" {
			meta = "User"
		}
		timeStr := ""
		if !m.CreateDate.IsZero() {
			timeStr = m.CreateDate.Local().Format("Jan 02, 15:04")
		}
		sb.WriteString(`<article class="sum-msg-card">`)
		sb.WriteString(`<header class="sum-msg-card-head">`)
		sb.WriteString(`<span class="sum-msg-author">` + template.HTMLEscapeString(meta) + `</span>`)
		if timeStr != "" {
			sb.WriteString(`<time class="sum-msg-time" datetime="` + template.HTMLEscapeString(m.CreateDate.Format(time.RFC3339)) + `">` + template.HTMLEscapeString(timeStr) + `</time>`)
		}
		sb.WriteString(`</header>`)
		sb.WriteString(`<div class="sum-msg-body">` + template.HTMLEscapeString(m.Body) + `</div>`)
		sb.WriteString(`</article>`)
	}
	if len(msgs) == 0 {
		sb.WriteString(`<div class="sum-msg-thread-empty" role="status">`)
		sb.WriteString(`<p class="sum-msg-thread-empty-title">No messages yet</p>`)
		sb.WriteString(`<p class="sum-msg-thread-empty-hint">Be the first to leave a comment on this record.</p>`)
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div>`)
	sb.WriteString(`<footer class="sum-msg-composer">`)
	sb.WriteString(`<form method="post" action="/web/chatter/post" class="sum-msg-form">`)
	sb.WriteString(`<input type="hidden" name="model" value="` + template.HTMLEscapeString(model) + `" />`)
	sb.WriteString(`<input type="hidden" name="res_id" value="` + template.HTMLEscapeString(fmt.Sprintf("%d", vr.RecordID)) + `" />`)
	sb.WriteString(`<input type="hidden" name="next" value="` + template.HTMLEscapeString(nextURL) + `" />`)
	sb.WriteString(`<label class="sr-only" for="sum-chatter-body">Message</label>`)
	sb.WriteString(`<textarea id="sum-chatter-body" name="body" rows="3" class="sum-msg-input" placeholder="Write a message…"></textarea>`)
	sb.WriteString(`<div class="sum-msg-form-actions"><button type="submit" class="sum-msg-send">Send</button></div>`)
	sb.WriteString(`</form></footer></div>`)
}

func formNewRecordURL(actionID int, menuID string) string {
	q := url.Values{}
	if actionID > 0 {
		q.Set("action", fmt.Sprintf("%d", actionID))
	}
	if strings.TrimSpace(menuID) != "" {
		q.Set("menu_id", strings.TrimSpace(menuID))
	}
	q.Set("view_type", "form")
	return "/web?" + q.Encode()
}

func RenderTree(view *parser.View, rows []map[string]interface{}, actionID int, menuID string, viewTabs []ViewSwitchTab) string {
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	var sb strings.Builder
	n := len(rows)
	listTitle := HumanViewBreadcrumb(view.Model, "tree")
	newHref := formNewRecordURL(actionID, menuID)

	sb.WriteString(`<div class="sum-tree-view">`)
	sb.WriteString(`<div class="sum-tree-control">`)
	sb.WriteString(`<div class="sum-tree-control-left">`)
	sb.WriteString(`<a href="` + template.HTMLEscapeString(newHref) + `" class="sum-tree-btn-new">New</a>`)
	sb.WriteString(`<h1 class="sum-tree-title">` + template.HTMLEscapeString(listTitle) + `</h1>`)
	sb.WriteString(`<button type="button" class="sum-tree-icon-btn" disabled aria-hidden="true" title="Configuration">`)
	sb.WriteString(`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>`)
	sb.WriteString(`</button>`)
	sb.WriteString(`</div>`)
	sb.WriteString(`<div class="sum-tree-search-wrap" role="search">`)
	sb.WriteString(`<span class="sum-tree-search-icon" aria-hidden="true"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="7"/><path d="M21 21l-4.35-4.35"/></svg></span>`)
	sb.WriteString(`<input type="search" class="sum-tree-search" placeholder="Search…" disabled aria-disabled="true" />`)
	sb.WriteString(`<span class="sum-tree-search-caret" aria-hidden="true"></span>`)
	sb.WriteString(`</div>`)
	sb.WriteString(`<div class="sum-tree-control-right">`)
	if n > 0 {
		sb.WriteString(fmt.Sprintf(`<span class="sum-tree-pager" aria-live="polite">1-%d / %d</span>`, n, n))
	} else {
		sb.WriteString(`<span class="sum-tree-pager">0 / 0</span>`)
	}
	sb.WriteString(`<div class="sum-tree-view-tabs" role="toolbar" aria-label="Switch view">`)
	for _, t := range viewTabs {
		active := ""
		if t.Active {
			active = " is-active"
		}
		sb.WriteString(`<a href="` + template.HTMLEscapeString(t.Href) + `" class="sum-tree-view-tab` + active + `" data-view="` + template.HTMLEscapeString(t.Mode) + `" title="` + template.HTMLEscapeString(t.Label) + `">`)
		if t.Mode == "kanban" {
			sb.WriteString(`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/></svg>`)
		} else if t.Mode == "tree" {
			sb.WriteString(`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75"><path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01"/></svg>`)
		} else {
			sb.WriteString(`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8l-6-6z"/><path d="M14 2v6h6M16 13H8M16 17H8M10 9H8"/></svg>`)
		}
		sb.WriteString(`</a>`)
	}
	sb.WriteString(`</div></div>`)

	sb.WriteString(`<div class="sum-web-table-wrap sum-tree-table-wrap">`)
	sb.WriteString(`<table class="sum-tree-table w-full text-left border-collapse">`)
	sb.WriteString(`<thead><tr>`)
	sb.WriteString(`<th class="sum-tree-th-check" scope="col"><span class="sum-tree-th-muted"><input type="checkbox" disabled aria-label="Select all" /></span></th>`)
	sb.WriteString(`<th class="sum-tree-th-grip" scope="col" aria-hidden="true"></th>`)
	for _, f := range view.Field {
		label := f.Label
		if label == "" {
			label = strings.Title(strings.ReplaceAll(f.Name, "_", " "))
		}
		sb.WriteString(`<th class="sum-tree-th">` + template.HTMLEscapeString(label) + `</th>`)
	}
	sb.WriteString(`<th class="sum-tree-th-actions" scope="col" aria-label="Columns"><button type="button" class="sum-tree-icon-btn sum-tree-icon-btn--ghost" disabled aria-hidden="true">`)
	sb.WriteString(`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 6h16M4 12h16M4 18h16"/></svg>`)
	sb.WriteString(`</button></th>`)
	sb.WriteString(`</tr></thead>`)
	sb.WriteString(`<tbody>`)
	colspan := len(view.Field) + 3
	if colspan < 3 {
		colspan = 3
	}
	if len(rows) == 0 {
		sb.WriteString(fmt.Sprintf(`<tr><td colspan="%d" class="sum-tree-empty">No records</td></tr>`, colspan))
	}
	for _, row := range rows {
		rid, hasID := orm.CoerceInt64(row["id"])
		menuTrim := strings.TrimSpace(menuID)
		canOpenForm := !view.TreeNoRowOpen && hasID && rid > 0 && (actionID > 0 || menuTrim != "")
		rowClass := "sum-tree-row"
		if canOpenForm {
			rowClass += " sum-tree-row--click"
		}
		sb.WriteString(`<tr class="` + rowClass + `"`)
		if canOpenForm {
			href := rowOpenURL(actionID, menuID, rid)
			qhref := strconv.Quote(href)
			sb.WriteString(` role="link" tabindex="0" onclick='window.location.href=` + qhref + `' onkeydown='if(event.key==="Enter"||event.key===" "){event.preventDefault();window.location.href=` + qhref + `}'`)
		}
		sb.WriteString(`>`)
		sb.WriteString(`<td class="sum-tree-td-check"><input type="checkbox" disabled onclick="event.stopPropagation()" aria-label="Select row" /></td>`)
		sb.WriteString(`<td class="sum-tree-td-grip" aria-hidden="true"><span class="sum-tree-grip">⠿</span></td>`)
		for _, f := range view.Field {
			cell := template.HTMLEscapeString(recStr(row, f.Name))
			sb.WriteString(`<td class="sum-tree-td">` + cell + `</td>`)
		}
		sb.WriteString(`<td class="sum-tree-td-actions"></td>`)
		sb.WriteString(`</tr>`)
	}
	sb.WriteString(`</tbody></table></div></div>`)

	return sb.String()
}

func RenderKanban(view *parser.View, rows []map[string]interface{}, actionID int, menuID string) string {
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	var sb strings.Builder

	sb.WriteString(`<div class="sum-kanban-board">`)
	sb.WriteString(`<div class="sum-kanban-columns">`)
	if len(rows) == 0 {
		sb.WriteString(`<div class="sum-kanban-empty">No records</div>`)
	}
	for _, row := range rows {
		rid, ok := orm.CoerceInt64(row["id"])
		if !ok {
			continue
		}
		href := rowOpenURL(actionID, menuID, rid)
		title := ""
		if len(view.Field) > 0 {
			title = recStr(row, view.Field[0].Name)
		}
		if title == "" {
			title = recStr(row, "name")
		}
		if title == "" {
			title = fmt.Sprintf("#%d", rid)
		}
		qhref := strconv.Quote(href)
		sb.WriteString(`<article class="sum-kanban-card" role="link" tabindex="0" onclick='window.location.href=` + qhref + `' onkeydown='if(event.key==="Enter"){window.location.href=` + qhref + `}'>`)
		sb.WriteString(`<h4 class="sum-kanban-card-title">` + template.HTMLEscapeString(title) + `</h4>`)
		var subParts []string
		for fi := 1; fi < len(view.Field); fi++ {
			s := recStr(row, view.Field[fi].Name)
			if s != "" {
				subParts = append(subParts, s)
			}
		}
		if len(subParts) > 0 {
			sb.WriteString(`<p class="sum-kanban-card-sub">` + template.HTMLEscapeString(strings.Join(subParts, " · ")) + `</p>`)
		}
		sb.WriteString(`</article>`)
	}
	sb.WriteString(`</div></div>`)

	return sb.String()
}

func RenderPivot(view *parser.View) string {
	var sb strings.Builder
	sb.WriteString(`<div class="bg-white p-8 rounded-lg shadow-sm border border-slate-200 flex flex-col items-center justify-center min-h-[400px]">`)
	sb.WriteString(`<svg class="w-16 h-16 text-slate-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg>`)
	sb.WriteString(`<h3 class="text-lg font-semibold text-slate-700">Pivot Table View</h3>`)
	sb.WriteString(`<p class="text-slate-500">Analytics and reporting will appear here.</p>`)
	sb.WriteString(`</div>`)
	return sb.String()
}

func renderGroup(sb *strings.Builder, g parser.Group, record map[string]interface{}, ro bool) {
	if ro {
		sb.WriteString(`<section class="sum-read-section">`)
		if g.Title != "" {
			sb.WriteString(`<h4 class="sum-read-section-title">` + template.HTMLEscapeString(strings.ToUpper(g.Title)) + `</h4>`)
		}
	} else {
		sb.WriteString(`<div class="group-container border border-slate-200 rounded-lg p-6 mb-6">`)
		if g.Title != "" {
			sb.WriteString(`<h4 class="text-sm font-bold uppercase tracking-wider text-slate-500 mb-6 pb-2 border-b border-slate-50">` + template.HTMLEscapeString(g.Title) + `</h4>`)
		}
	}
	for _, sep := range g.Separator {
		renderSeparator(sb, sep)
	}
	for _, lab := range g.Label {
		renderLabel(sb, lab)
	}
	if ro {
		sb.WriteString(`<div class="sum-read-fields">`)
	} else {
		sb.WriteString(`<div class="grid grid-cols-1 md:grid-cols-2 gap-6">`)
	}
	for _, f := range g.Field {
		renderField(sb, f, record, ro)
	}
	for _, subG := range g.Group {
		renderGroup(sb, subG, record, ro)
	}
	sb.WriteString(`</div>`)
	if ro {
		sb.WriteString(`</section>`)
	} else {
		sb.WriteString(`</div>`)
	}
}

func rawField(record map[string]interface{}, name string) (interface{}, bool) {
	if record == nil {
		return nil, false
	}
	v, ok := record[name]
	return v, ok
}

func renderField(sb *strings.Builder, f parser.Field, record map[string]interface{}, ro bool) {
	if gs := strings.TrimSpace(f.Groups); gs != "" {
		if !orm.UserHasAnyAccessGroup(orm.SecurityUID(), gs) {
			return
		}
	}
	label := f.Label
	if label == "" {
		label = strings.Title(strings.ReplaceAll(f.Name, "_", " "))
	}

	raw, hasRaw := rawField(record, f.Name)
	isBoolish := f.Widget == "boolean" || strings.HasSuffix(f.Name, "_active")
	if isBoolish {
		checked := ""
		if hasRaw && isTruthyDB(raw) {
			checked = ` checked`
		}
		dis := ""
		if ro {
			dis = ` disabled`
		}
		if ro {
			val := "No"
			if hasRaw && isTruthyDB(raw) {
				val = "Yes"
			}
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value sum-read-value--bool">` + template.HTMLEscapeString(val) + `</div>`)
			sb.WriteString(`</div>`)
			return
		}
		sb.WriteString(`<div class="o_field_widget flex flex-col space-y-1">`)
		sb.WriteString(`<label class="text-xs font-bold text-slate-500 uppercase tracking-wide" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
		sb.WriteString(fmt.Sprintf(`<input class="h-4 w-4 rounded border-slate-300 text-indigo-600" id="%s" name="%s" type="checkbox"%s%s />`,
			template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name), checked, dis))
		sb.WriteString(`</div>`)
		return
	}

	if f.Widget == "image" {
		src := recStr(record, f.Name)
		if ro {
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value">`)
			if src != "" && (strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:")) {
				sb.WriteString(fmt.Sprintf(`<img src="%s" alt="" class="max-h-24 rounded border border-slate-200" />`, template.HTMLEscapeString(src)))
			} else {
				sb.WriteString(`<span class="text-slate-400">—</span>`)
			}
			sb.WriteString(`</div></div>`)
			return
		}
		sb.WriteString(`<div class="o_field_widget flex flex-col space-y-1">`)
		if src != "" && (strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:")) {
			sb.WriteString(`<div class="w-24 h-24 border border-slate-200 rounded-sm overflow-hidden">`)
			sb.WriteString(fmt.Sprintf(`<img src="%s" alt="" class="w-full h-full object-cover" />`, template.HTMLEscapeString(src)))
			sb.WriteString(`</div>`)
		} else {
			sb.WriteString(`<div class="w-24 h-24 bg-slate-50 border border-slate-200 rounded-sm flex items-center justify-center">`)
			sb.WriteString(`<svg class="w-8 h-8 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg>`)
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
	} else if f.Widget == "many2many_tags" {
		txt := recStr(record, f.Name)
		if ro {
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value">` + template.HTMLEscapeString(txt) + `</div>`)
			sb.WriteString(`</div>`)
			return
		}
		sb.WriteString(`<div class="o_field_widget flex flex-col space-y-1">`)
		sb.WriteString(`<label class="text-xs font-bold text-slate-500 uppercase tracking-wide" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
		sb.WriteString(`<div class="flex flex-wrap gap-1 p-2 border border-slate-200 rounded-sm bg-slate-50 min-h-[38px] text-sm text-slate-700">`)
		sb.WriteString(template.HTMLEscapeString(txt))
		sb.WriteString(`</div></div>`)
	} else {
		placeholder := f.Placeholder
		if placeholder == "" {
			placeholder = "Enter " + strings.ToLower(label) + "..."
		}
		val := recStr(record, f.Name)
		if ro {
			display := val
			if display == "" {
				display = "—"
			}
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value">` + template.HTMLEscapeString(display) + `</div>`)
			sb.WriteString(`</div>`)
			return
		}
		sb.WriteString(`<div class="o_field_widget flex flex-col space-y-1">`)
		sb.WriteString(`<label class="text-xs font-bold text-slate-500 uppercase tracking-wide" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
		sb.WriteString(fmt.Sprintf(`<input class="px-3 py-2 border border-slate-200 rounded-sm text-sm focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all" id="%s" name="%s" type="text" placeholder="%s" value="%s" />`,
			template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name),
			template.HTMLEscapeString(placeholder), template.HTMLEscapeString(val)))
		sb.WriteString(`</div>`)
	}
}

func writeResUsersSecuritySection(sb *strings.Builder, vr *ViewRecordData, ro bool) {
	uid := orm.SecurityUID()
	if uid <= 0 {
		return
	}
	if err := orm.CheckModelAccess(uid, "res.groups", "read"); err != nil {
		return
	}
	sb.WriteString(`<div class="border border-slate-200 rounded-lg p-6 mb-6 bg-slate-50/80">`)
	sb.WriteString(`<h3 class="text-sm font-semibold text-slate-800 mb-3">Access rights</h3>`)
	if !ro {
		sb.WriteString(`<input type="hidden" name="security_groups_touched" value="1"/>`)
		sb.WriteString(`<div class="o_field_widget flex flex-col space-y-1 mb-4">`)
		sb.WriteString(`<label class="text-xs font-bold text-slate-500 uppercase tracking-wide" for="password_plain">New password</label>`)
		sb.WriteString(`<input class="px-3 py-2 border border-slate-200 rounded-sm text-sm" id="password_plain" name="password_plain" type="password" autocomplete="new-password" placeholder="Leave blank to keep current" />`)
		sb.WriteString(`</div>`)
	}
	selected := map[int]struct{}{}
	if vr.RecordID > 0 {
		rel := orm.GetTableName("res.groups.user.rel")
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
	groups, err := orm.ListAllGroupRows()
	if err != nil || len(groups) == 0 {
		sb.WriteString(`<p class="text-sm text-slate-500">No groups defined.</p></div>`)
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
		sb.WriteString(`<p class="text-sm text-slate-700">` + template.HTMLEscapeString(strings.Join(names, ", ")) + `</p></div>`)
		return
	}
	sb.WriteString(`<div class="space-y-2 max-h-48 overflow-y-auto">`)
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
		sb.WriteString(`<label class="flex items-center gap-2 text-sm text-slate-700">`)
		sb.WriteString(fmt.Sprintf(`<input type="checkbox" name="security_group_ids" value="%d"%s />`, int(gid), chk))
		sb.WriteString(template.HTMLEscapeString(nm))
		sb.WriteString(`</label>`)
	}
	sb.WriteString(`</div></div>`)
}
