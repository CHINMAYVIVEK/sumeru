# Web layer

The browser UI is **server-rendered HTML** plus static **CSS** and **JS** under **`core/engine/assets/`**. There is no SPA framework requirement for core workspace pages.

## Handlers and templates

HTTP handlers under **`core/server/web/`** build **`render.PageData`** and execute Go **`html/template`** files from **`core/engine/templates/`** (`base.html`, inner layouts).

## Render package (`core/engine/render/`)

| Area | Files (indicative) |
| ---- | ------------------- |
| Dispatch | `view_render.go` (`RenderView`), `form_render.go` (`RenderForm`) |
| Lists / pivots | `tree_render.go`, `kanban_render.go`, `pivot_render.go` |
| Workspace chrome | `workspace_chrome.go` |
| Chatter | `activity_chatter_render.go` |
| User security tab | `user_security_render.go` |
| Hooks / types | `render_types.go` — **`RegisterShellHook`**, **`RegisterNotebookHook`** |

Hooks let addons inject HTML **without** forking templates: shell chrome and per-notebook-page bodies (matched by model + page title).

## Assets

- Global CSS order is defined in **`core/engine/assets/stylesheets.go`**.
- Per-page extra sheets go through **`PageData.ViewStylesheetURLs`** (Apps grid, Home hub, Settings hub, etc.).
- Optional per-addon **`static/css/theme-overrides.css`** is exposed as **`/static/addon-css/<module>.css`**.

Details and file names: **`sumeru/README.md`** (UI section).

## Client scripts

**`core/engine/assets/js/core.js`** wires sidebar, workspace navigation, activity panel, and form save/cancel. Feature-specific snippets (e.g. messages composer) live beside it under **`assets/js/ui/`**.
