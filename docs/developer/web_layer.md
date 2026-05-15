# Web layer

The browser UI is **server-rendered HTML** plus static **CSS** and **JS** under **`core/engine/assets/`**. There is no SPA framework requirement for core workspace pages.

## Handlers and templates

HTTP handlers under **`core/server/web/`** build **`render.PageData`** and execute Go **`html/template`** files from **`core/engine/templates/`** (`base.html`, inner layouts).

## Render package (`core/engine/render/`)

Import rules live in **`core/engine/render/doc.go`**. File-level layout (which `.go` owns forms, tree, hooks, etc.) is documented in **`sumeru/README.md`** — keep that table in README only.

Hooks let addons inject HTML **without** forking templates: shell chrome and per-notebook-page bodies (matched by model + page title).

## Assets

- Register a new **global** stylesheet in **`core/engine/assets/stylesheets.go`** (`DefaultStylesheetURLs`) and document it in **`sumeru/README.md`** (UI section). Do not copy the core CSS file list into this doc.
- Per-page extra sheets go through **`PageData.ViewStylesheetURLs`** (Apps grid, Home hub, Settings hub, etc.).
- Optional per-addon **`static/css/theme-overrides.css`** is exposed as **`/static/addon-css/<module>.css`**.

More detail: **`sumeru/README.md`** (UI section).

## Client scripts

**`core/engine/assets/js/core.js`** wires sidebar, workspace navigation, activity panel, and form save/cancel. Feature-specific snippets (e.g. messages composer) live beside it under **`assets/js/ui/`**.
