package render

import (
	"fmt"
	"html/template"
	"log"
	"net/url"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

type PageData struct {
	Title               string // legacy / diagnostics; prefer ViewBreadcrumb for UI
	ViewBreadcrumb      string // human label for breadcrumb (Odoo-style; never technical model id)
	AppName             string // product display name (browser tab suffix, header)
	ModuleName          string
	Content             template.HTML
	TopMenus            []parser.MenuItem
	SidebarMenus        []SidebarMenu
	ActiveModuleID      string
	ActiveMenuID        string
	ViewStylesheetURLs  []string
	AppsNavActive       bool
	ExtraStylesheetURLs []string
	LogoURL             string
	ShellCompany        string
	ShellUser           string
	UserInitial         string // first letter for avatar
}

type SidebarMenu struct {
	ID       string
	Name     string
	SubMenus []parser.MenuItem
}

// ViewRecordData carries rows loaded from the ORM for HTML rendering.
type ViewRecordData struct {
	ActionID int
	Record   map[string]interface{}
	ListRows []map[string]interface{}
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
	q.Set("action", fmt.Sprintf("%d", actionID))
	if strings.TrimSpace(menuID) != "" {
		q.Set("menu_id", strings.TrimSpace(menuID))
	}
	q.Set("id", fmt.Sprintf("%d", rowID))
	return "/web?" + q.Encode()
}

func RenderView(view *parser.View, activeMenuID, templatesDir string, recData *ViewRecordData) string {
	if recData == nil {
		recData = &ViewRecordData{}
	}
	var content string
	switch view.Type {
	case "form":
		content = RenderForm(view, recData.Record)
	case "tree", "list":
		content = RenderTree(view, recData.ListRows, recData.ActionID, activeMenuID)
	case "kanban":
		content = RenderKanban(view, recData.ListRows, recData.ActionID, activeMenuID)
	case "pivot":
		content = RenderPivot(view)
	default:
		content = RenderForm(view, recData.Record)
	}

	topMenus, sidebarMenus, activeModuleID := LoadShellMenus(activeMenuID)
	moduleName := ModuleNameForTopMenu(topMenus, activeModuleID)
	viewBC := HumanViewBreadcrumb(view.Model, view.Type)

	pageData := PageData{
		Title:               fmt.Sprintf("%s · %s", view.Model, view.Type),
		ViewBreadcrumb:      viewBC,
		ModuleName:          moduleName,
		Content:             template.HTML(content),
		TopMenus:            topMenus,
		SidebarMenus:        sidebarMenus,
		ActiveModuleID:      activeModuleID,
		ActiveMenuID:        activeMenuID,
		ViewStylesheetURLs:  []string{"/static/css/view-web.css"},
		ExtraStylesheetURLs: ExtraStylesheetURLs,
	}

	log.Printf("Rendering view for model %s (ActiveMenu: %s, ActiveModule: %s)", view.Model, activeMenuID, activeModuleID)

	out, err := RenderPage(templatesDir, pageData)
	if err != nil {
		log.Printf("Error rendering page: %v", err)
		return content
	}
	return out
}

func RenderForm(view *parser.View, record map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString(`<div class="o_form_view flex h-full overflow-hidden">`)

	// Main Content Area
	sb.WriteString(`<div class="o_form_sheet_bg flex-1 bg-slate-100 overflow-y-auto p-4 md:p-8">`)

	// Header (Buttons & Status)
	if view.Header != nil {
		renderHeader(&sb, view.Header, record)
	}

	// Sheet (The actual form content)
	if view.Sheet != nil {
		renderSheet(&sb, view.Sheet, record)
	} else {
		// Fallback for simple views
		sb.WriteString(`<div class="o_form_sheet bg-white shadow-sm border border-slate-200 rounded-sm p-8 min-h-full">`)
		for _, f := range view.Field {
			renderField(&sb, f, record)
		}
		for _, g := range view.Group {
			renderGroup(&sb, g, record)
		}
		sb.WriteString(`</div>`)
	}

	if view.Footer != nil {
		renderFormFooter(&sb, view.Footer)
	}

	sb.WriteString(`</div>`)

	// Chatter (Right Sidebar)
	if view.Chatter != nil {
		renderChatter(&sb, view.Chatter)
	}

	sb.WriteString(`</div>`)

	return sb.String()
}

func renderHeader(sb *strings.Builder, h *parser.Header, record map[string]interface{}) {
	sb.WriteString(`<div class="o_form_statusbar flex items-center justify-between bg-white border border-slate-200 rounded-t-sm p-2 mb-0 border-b-0">`)

	// Buttons
	sb.WriteString(`<div class="o_statusbar_buttons flex space-x-2">`)
	for _, b := range h.Button {
		class := "px-3 py-1.5 rounded text-sm font-bold shadow-sm transition-all "
		if b.Class == "oe_highlight" {
			class += "bg-indigo-600 text-white hover:bg-indigo-700"
		} else {
			class += "bg-white text-slate-700 border border-slate-300 hover:bg-slate-50"
		}
		sb.WriteString(fmt.Sprintf(`<button type="button" class="%s">%s</button>`, class, template.HTMLEscapeString(b.String)))
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

func renderSheet(sb *strings.Builder, s *parser.Sheet, record map[string]interface{}) {
	sb.WriteString(`<div class="o_form_sheet bg-white shadow-sm border border-slate-200 rounded-b-sm p-8 min-h-full">`)

	// Render Stat Buttons if any (Usually inside a div with class oe_button_box)
	for _, d := range s.Div {
		if strings.Contains(d.Class, "oe_button_box") {
			renderButtonBox(sb, d)
		}
	}

	// Render Title/Avatar area
	for _, d := range s.Div {
		if strings.Contains(d.Class, "oe_title") {
			renderTitle(sb, d, record)
		}
	}

	// Render direct fields/groups in sheet
	sb.WriteString(`<div class="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">`)
	for _, sep := range s.Separator {
		renderSeparator(sb, sep)
	}
	for _, lab := range s.Label {
		renderLabel(sb, lab)
	}
	for _, g := range s.Group {
		renderGroup(sb, g, record)
	}
	for _, f := range s.Field {
		renderField(sb, f, record)
	}
	sb.WriteString(`</div>`)

	// Render Notebooks (Tabs)
	for _, nb := range s.Notebook {
		renderNotebook(sb, nb, record)
	}

	sb.WriteString(`</div>`)
}

func renderButtonBox(sb *strings.Builder, d parser.Div) {
	_ = d
	sb.WriteString(`<div class="flex justify-end mb-8 -mr-8 -mt-8 border-b border-slate-100 min-h-[1px]"></div>`)
}

func renderTitle(sb *strings.Builder, d parser.Div, record map[string]interface{}) {
	sb.WriteString(`<div class="flex items-start space-x-6 mb-8">`)

	sb.WriteString(`<div class="relative group">`)
	sb.WriteString(`<div class="w-32 h-32 bg-slate-100 rounded-sm border-2 border-dashed border-slate-200 flex items-center justify-center overflow-hidden">`)
	sb.WriteString(`<svg class="w-12 h-12 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>`)
	sb.WriteString(`</div></div>`)

	sb.WriteString(`<div class="flex-1">`)
	for _, h1 := range d.H1 {
		for _, f := range h1.Field {
			ph := template.HTMLEscapeString(f.Label)
			v := recStr(record, f.Name)
			sb.WriteString(fmt.Sprintf(`<input class="text-4xl font-bold text-slate-800 border-b border-transparent hover:border-slate-200 focus:border-indigo-500 focus:outline-none w-full bg-transparent pb-1 mb-4" placeholder="%s" name="%s" value="%s" readonly />`,
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

		sb.WriteString(`<div class="flex items-center text-slate-500 group">`)
		sb.WriteString(fmt.Sprintf(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="%s"></path></svg>`, icon))
		v := recStr(record, f.Name)
		sb.WriteString(fmt.Sprintf(`<input class="text-sm border-b border-transparent hover:border-slate-200 focus:border-indigo-500 focus:outline-none bg-transparent py-0.5" placeholder="%s" name="%s" value="%s" readonly />`,
			template.HTMLEscapeString(f.Label), template.HTMLEscapeString(f.Name), template.HTMLEscapeString(v)))
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div>`)

	sb.WriteString(`</div>`)
	sb.WriteString(`</div>`)
}

func renderNotebook(sb *strings.Builder, nb parser.Notebook, record map[string]interface{}) {
	sb.WriteString(`<div class="o_notebook mt-12">`)

	// Tabs Header
	sb.WriteString(`<div class="flex border-b border-slate-200 mb-6">`)
	for i, p := range nb.Page {
		activeClass := ""
		if i == 0 {
			activeClass = "border-indigo-500 text-indigo-700 bg-white"
		} else {
			activeClass = "border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300"
		}
		sb.WriteString(fmt.Sprintf(`<button type="button" class="px-6 py-3 border-b-2 font-bold text-sm transition-all %s">%s</button>`, activeClass, template.HTMLEscapeString(p.Title)))
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
			renderGroup(sb, g, record)
		}
		for _, f := range p.Field {
			renderField(sb, f, record)
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
		sb.WriteString(fmt.Sprintf(`<button type="button" name="%s" class="%s">%s</button>`,
			template.HTMLEscapeString(b.Name), cls, label))
	}
	sb.WriteString(`</div>`)
}

func renderChatter(sb *strings.Builder, c *parser.Chatter) {
	_ = c
	sb.WriteString(`<div class="o_form_chatter w-[400px] border-l border-slate-200 bg-white flex flex-col h-full shadow-xl">`)
	sb.WriteString(`<div class="p-4 border-b border-slate-100 flex space-x-2 bg-slate-50/50">`)
	sb.WriteString(`<span class="text-xs font-semibold text-slate-500 uppercase tracking-wide">Messages</span>`)
	sb.WriteString(`</div>`)
	sb.WriteString(`<div class="flex-1 overflow-y-auto p-6 text-sm text-slate-500">`)
	sb.WriteString(`</div>`)
	sb.WriteString(`</div>`)
}

func RenderTree(view *parser.View, rows []map[string]interface{}, actionID int, menuID string) string {
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	var sb strings.Builder

	sb.WriteString(`<div class="bg-white shadow-sm rounded-sm border border-slate-200 overflow-hidden">`)
	sb.WriteString(`<table class="w-full text-left border-collapse">`)
	sb.WriteString(`<thead class="bg-slate-50 border-b border-slate-200">`)
	sb.WriteString(`<tr>`)
	for _, f := range view.Field {
		label := f.Label
		if label == "" {
			label = strings.Title(strings.ReplaceAll(f.Name, "_", " "))
		}
		sb.WriteString(`<th class="px-4 py-3 text-xs font-bold text-slate-500 uppercase tracking-wider">` + template.HTMLEscapeString(label) + `</th>`)
	}
	sb.WriteString(`</tr>`)
	sb.WriteString(`</thead>`)
	sb.WriteString(`<tbody class="divide-y divide-slate-100">`)
	colspan := len(view.Field)
	if colspan < 1 {
		colspan = 1
	}
	if len(rows) == 0 {
		sb.WriteString(fmt.Sprintf(`<tr><td colspan="%d" class="px-4 py-8 text-sm text-slate-500 text-center">No records</td></tr>`, colspan))
	}
	for _, row := range rows {
		rid, ok := orm.CoerceInt64(row["id"])
		if !ok {
			continue
		}
		href := rowOpenURL(actionID, menuID, rid)
		sb.WriteString(`<tr class="hover:bg-slate-50 transition-colors cursor-pointer" onclick="window.location.href=` + strconv.Quote(href) + `">`)
		for _, f := range view.Field {
			cell := template.HTMLEscapeString(recStr(row, f.Name))
			sb.WriteString(`<td class="px-4 py-3 text-sm text-slate-600">` + cell + `</td>`)
		}
		sb.WriteString(`</tr>`)
	}
	sb.WriteString(`</tbody>`)
	sb.WriteString(`</table>`)
	sb.WriteString(`</div>`)

	return sb.String()
}

func RenderKanban(view *parser.View, rows []map[string]interface{}, actionID int, menuID string) string {
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	var sb strings.Builder

	sb.WriteString(`<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">`)
	if len(rows) == 0 {
		sb.WriteString(`<div class="col-span-full text-sm text-slate-500 py-8 text-center">No records</div>`)
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
		sb.WriteString(`<div class="bg-white p-4 rounded-lg shadow-sm border border-slate-200 hover:shadow-md transition-shadow cursor-pointer" onclick="window.location.href=` + strconv.Quote(href) + `">`)
		sb.WriteString(`<h4 class="font-bold text-slate-800">` + template.HTMLEscapeString(title) + `</h4>`)
		if len(view.Field) > 1 {
			sb.WriteString(`<p class="text-sm text-slate-500 mt-2">` + template.HTMLEscapeString(recStr(row, view.Field[1].Name)) + `</p>`)
		}
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div>`)

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

func renderGroup(sb *strings.Builder, g parser.Group, record map[string]interface{}) {
	sb.WriteString(`<div class="group-container border border-slate-200 rounded-lg p-6 mb-6">`)
	if g.Title != "" {
		sb.WriteString(`<h4 class="text-sm font-bold uppercase tracking-wider text-slate-500 mb-6 pb-2 border-b border-slate-50">` + template.HTMLEscapeString(g.Title) + `</h4>`)
	}
	for _, sep := range g.Separator {
		renderSeparator(sb, sep)
	}
	for _, lab := range g.Label {
		renderLabel(sb, lab)
	}
	sb.WriteString(`<div class="grid grid-cols-1 md:grid-cols-2 gap-6">`)
	for _, f := range g.Field {
		renderField(sb, f, record)
	}
	for _, subG := range g.Group {
		renderGroup(sb, subG, record)
	}
	sb.WriteString(`</div>`)
	sb.WriteString(`</div>`)
}

func rawField(record map[string]interface{}, name string) (interface{}, bool) {
	if record == nil {
		return nil, false
	}
	v, ok := record[name]
	return v, ok
}

func renderField(sb *strings.Builder, f parser.Field, record map[string]interface{}) {
	label := f.Label
	if label == "" {
		label = strings.Title(strings.ReplaceAll(f.Name, "_", " "))
	}

	sb.WriteString(`<div class="o_field_widget flex flex-col space-y-1">`)
	sb.WriteString(`<label class="text-xs font-bold text-slate-500 uppercase tracking-wide" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)

	raw, hasRaw := rawField(record, f.Name)
	isBoolish := f.Widget == "boolean" || strings.HasSuffix(f.Name, "_active")
	if isBoolish {
		checked := ""
		if hasRaw && isTruthyDB(raw) {
			checked = ` checked`
		}
		sb.WriteString(fmt.Sprintf(`<input class="h-4 w-4 rounded border-slate-300 text-indigo-600" id="%s" name="%s" type="checkbox"%s disabled />`,
			template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name), checked))
		sb.WriteString(`</div>`)
		return
	}

	if f.Widget == "image" {
		src := recStr(record, f.Name)
		if src != "" && (strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:")) {
			sb.WriteString(`<div class="w-24 h-24 border border-slate-200 rounded-sm overflow-hidden">`)
			sb.WriteString(fmt.Sprintf(`<img src="%s" alt="" class="w-full h-full object-cover" />`, template.HTMLEscapeString(src)))
			sb.WriteString(`</div>`)
		} else {
			sb.WriteString(`<div class="w-24 h-24 bg-slate-50 border border-slate-200 rounded-sm flex items-center justify-center">`)
			sb.WriteString(`<svg class="w-8 h-8 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg>`)
			sb.WriteString(`</div>`)
		}
	} else if f.Widget == "many2many_tags" {
		txt := recStr(record, f.Name)
		sb.WriteString(`<div class="flex flex-wrap gap-1 p-2 border border-slate-200 rounded-sm bg-slate-50 min-h-[38px] text-sm text-slate-700">`)
		sb.WriteString(template.HTMLEscapeString(txt))
		sb.WriteString(`</div>`)
	} else {
		placeholder := f.Placeholder
		if placeholder == "" {
			placeholder = "Enter " + strings.ToLower(label) + "..."
		}
		val := recStr(record, f.Name)
		sb.WriteString(fmt.Sprintf(`<input class="px-3 py-2 border border-slate-200 rounded-sm text-sm focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all" id="%s" name="%s" type="text" placeholder="%s" value="%s" readonly />`,
			template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name),
			template.HTMLEscapeString(placeholder), template.HTMLEscapeString(val)))
	}
	sb.WriteString(`</div>`)
}
