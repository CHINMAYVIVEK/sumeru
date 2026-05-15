# Sumeru

**Sumeru** is an experimental ERP-style web application in Go (this repository’s Go module is **`sumeru`**): PostgreSQL-backed ORM, **addons** (module XML under `<sumeru>` + manifests), model sync on startup, and a web UI (XML views, plain CSS, shell with sidebar and activity panel).

- **Entry binary:** `cmd/sumeru/main.go` → **`sumeru/core/server`** (`server.Run`). Library code under **`core/`** has no `main`; addon code uses **`sumeru/core/base`** as the stable API.
- **Addon Go code:** import **`sumeru/core/base`** for models (stable API); avoid importing **`sumeru/core/orm`** directly in new code.

**Documentation:** **[`docs/developer/index.md`](docs/developer/index.md)** (concepts and reference layout) · **[`docs/howtos/index.md`](docs/howtos/index.md)** (step-by-step recipes) · user intro in **`docs/user/`** · RST stubs in **`docs/refrence/`**.

**Convention:** path-related INI values resolve from the **INI file’s directory** (and optional **`sumeru_home`** for the **`core/base`** tree). When running only from the **`sumeru`** repo root with **`sumeru.conf`** there, that matches the old “cwd = repo root” habit; prefer absolute paths in production.

---

## Prerequisites

| Requirement | Use case |
| ----------- | -------- |
| [Go](https://go.dev/dl/) **1.26.2+** (see `go.mod`) | Build and `go run` the server |
| [PostgreSQL](https://www.postgresql.org/) | Application database |

---

## One-time setup

| Step | Command / action | Use case |
| ---- | ---------------- | -------- |
| 1. Create DB | `psql … -c "CREATE DATABASE sumeru;"` (or your GUI) | Empty database matching `db_name` in `sumeru.conf` |
| 2. Configure | Edit **`sumeru.conf`** | Point `db_*`, `http_port`, `addons_path` at your environment |
| 3. Download modules | `go mod download` | Populate the Go module cache (CI or fresh clone) |

---

## Standard `sumeru` + workspace `sumeru_custom_addons`

Keep the **`sumeru`** Go module pristine for `git pull` (no local churn from custom addon sets). Use the sibling **`sumeru_custom_addons`** workspace (same parent directory as **`sumeru/`** in this monorepo) to run the server with your own INI and a generated blank-import file that lives **outside** **`sumeru/`**:

| Piece | Role |
| ----- | ---- |
| **`sumeru/`** | Upstream-style standard code: `go generate ./cmd/sumeru` refreshes **`cmd/sumeru/zimports.go`** when you develop entirely inside this module. |
| **`sumeru_custom_addons/`** | Separate `go.mod` with `replace sumeru => ../sumeru`, thin **`main.go`**, and **`make generate`** writing **`addonimports/zimports.go`** via **`sumeru-import-gen`** (`-root ../sumeru`, **`-package addonimports`**, **`-out`** absolute or relative to **`-root`**). |

From **`sumeru_custom_addons`**: copy **`sumeru.conf.example`** to **`sumeru.conf`**, then **`make run`**. Details and custom-addon import constraints: **[`../sumeru_custom_addons/README.md`](../sumeru_custom_addons/README.md)**.

**`sumeru-import-gen`** flags (also used by `go generate ./cmd/sumeru` for the default case):

| Flag | Default | Use case |
| ---- | ------- | -------- |
| **`-root`** | cwd walk-up | Standard **`sumeru`** repo root (directory with **`go.mod`**). |
| **`-config`** | `sumeru.conf` | INI path; absolute, or relative to **`-root`**. |
| **`-out`** | `cmd/sumeru/zimports.go` | Output file; if **relative**, resolved against **`-root`** (use an **absolute** path to write under **`sumeru_custom_addons`**). |
| **`-package`** | `main` | Generated file’s package name (e.g. **`addonimports`** for the workspace). |

**Custom Go addons** in **`sumeru_custom_addons/addons/`** ship as **`sumeru_custom_addons/addons/<name>`** in generated imports; addons under the standard tree stay **`sumeru/…`**. Symlinking into **`sumeru/addons/`** is still valid if you prefer a single module path.

---

## Configuration (`sumeru.conf`)

INI format: `key = value` under **`[options]`**. Lines starting with `#` or `;` are comments.

| Key | Use case |
| --- | -------- |
| `db_host`, `db_port` | PostgreSQL host and port |
| `db_user`, `db_password` | Database credentials |
| `db_name` | Database name (overridable at runtime with **`-d`** / **`--database`**) |
| `db_sslmode` | PostgreSQL SSL mode (e.g. `disable` for local dev) |
| `http_port` | HTTP listen port (default **8080**; overridable with **`--http-port`** / **`-p`**) |
| `addons_path` | Comma-separated addon roots; **`core/base`** (platform **`base`**, **`user`**, **`company`**) is always prepended first. Later roots **override** duplicate module **names**. Relative segments are resolved from the **INI file’s directory**. |
| `sumeru_home` | Optional. If set, **`core/base`** loads from this directory under the standard **`sumeru`** checkout. Relative values are resolved from the **INI file’s directory**. When set, default **`assets_path`** / **`templates_path`** (if omitted in the INI) are under this tree. Use for configs outside the repo (e.g. **`sumeru_custom_addons/sumeru.conf`**). |
| `assets_path` | Static files root (default **`core/engine/assets`**; resolved from INI dir unless absolute; see **`sumeru_home`**) |
| `templates_path` | HTML templates (default **`core/engine/templates`**; same resolution rules) |
| `logo_path` | Optional image file; served at **`/static/app-logo`** (`.svg`, `.png`, `.jpg`/`.jpeg`, `.webp`) |
| `company_display_name` | Optional header chip; if empty and **`company`** module is installed, first **`core.company`** name is used |
| `user_display_name` | Optional header label; if empty and **`user`** module is installed, first **`core.user`** display is used |
| `brand_css` | Optional extra CSS file; linked as **`/static/brand.css`** after view stylesheets |
| `log_file` | Optional; if set, **`log`** output is **appended** to this file and still written to **stderr**; parent directories are created; path is absolutized like other paths |

---

## CLI (`server.Run`)

`server.Run` (all entrypoints below) accepts the flags below in any order; **comma-separated** lists install or update **multiple** modules in one run.

| Flag / pattern | Use case |
| ---------------- | -------- |
| **`-c path`** | INI config file |
| **`-d db`** / **`--database db`** | Overrides **`db_name`** in the INI for this process only |
| **`-i mod`** or **`-i mod1,mod2`** | Install one or many modules |
| **`-u mod`**, **`-u mod1,mod2`**, or **`-u all`** | Update / reload metadata; **`all`** = every **`sys.module`** row with **`state = installed`** (dependency order on disk) |
| **`--http-port N`** / **`-p N`** | Overrides **`http_port`** in the INI |
| **`--stop-after-init`** | After **`-i`** / **`-u`**, exit without starting HTTP |

**Precedence:** **`--database`** wins over **`-d`** if both are set; **`-p`** wins over **`--http-port`** if both are set.

---

### Start the HTTP server (normal development)

| Command | Use case |
| ------- | -------- |
| `make run` | `go run` with default **`-c sumeru.conf`** |
| `make run EXTRA_RUN_FLAGS='-p 9090'` | Same, listen on **9090** |
| `./sumeru.sh` | No args → **`-c sumeru.conf`** (or **`SUMERU_CONF`**) |
| `go run ./cmd/sumeru -- -c sumeru.conf -p 9090` | Explicit config + custom port |
| `go run ./cmd/sumeru -- -c sumeru.conf --http-port 9090 -d sumeru_dev` | Config + port + database override |
| `./sumeru -c sumeru.conf -p 9090` | After **`make build`** |

Then open **`http://localhost:<port>`** — `/` redirects to **`/web/apps`**. After sign-in, the shell includes **Home** (`/web/home`), **Settings** (`/web/settings` — configuration sections and installed apps), **Apps**, and top-level module menus; data views open as **`/web?…`** (tree, form, kanban).

---

### Use a different config file, database, or port

| Command | Use case |
| ------- | -------- |
| `go run ./cmd/sumeru -- -c /path/to/other.conf` | Alternate INI |
| `./sumeru.sh -c /path/to/other.conf` | Same via script |
| `SUMERU_CONF=/path/to/other.conf ./sumeru.sh` | Script injects **`-c`** from env when you pass no args |
| `go run ./cmd/sumeru -- -c sumeru.conf -d other_db` | Use PostgreSQL database **`other_db`** |
| `go run ./cmd/sumeru -- -c sumeru.conf --database other_db` | Same (**`--database`**) |
| `go run ./cmd/sumeru -- -c sumeru.conf -p 9090` | Listen on **9090** regardless of INI **`http_port`** |

---

### Install modules (first time or new apps)

| Command | Use case |
| ------- | -------- |
| `go run ./cmd/sumeru -- -c sumeru.conf -i sales -p 9090` | Install **sales**, serve on **9090** |
| `go run ./cmd/sumeru -- -c sumeru.conf -i sales,crm -p 9090 -d sumeru` | **Multiple** modules + custom port + database |
| `./sumeru.sh -i company,user --stop-after-init` | Install **company** and **user**, then **exit** (CI / bootstrap) |
| `go run ./cmd/sumeru -- -c sumeru.conf -i sales,sale_demo_inherit --stop-after-init -p 8080` | Install selected modules **without** keeping HTTP (port still parsed; server not started) |

**Typical first apps:** `sales`, `crm`, `inventory`, **`company`**, **`user`**, `sale_demo_inherit` (as needed).

---

### Update modules (reload XML / metadata from disk)

| Command | Use case |
| ------- | -------- |
| `go run ./cmd/sumeru -- -c sumeru.conf -u all -p 9090 --stop-after-init` | Reload **every installed** module on disk; rows missing from **`addons_path`** are skipped, then exit |
| `go run ./cmd/sumeru -- -c sumeru.conf -u sales -p 9090` | Refresh **sales** on port **9090** |
| `go run ./cmd/sumeru -- -c sumeru.conf -u sales,sale_demo_inherit -p 9090 -d sumeru` | **Multiple** modules + port + database |
| `./sumeru.sh -u sales,sale_demo_inherit --stop-after-init` | Batch update then exit |
| `make run EXTRA_RUN_FLAGS='-u sales -p 9090'` | Update via **`make`** then serve (same invocation runs **`-u`** then server) |

Use **`-u`** after changing **`views/*.xml`**, **`menus.xml`**, or **`manifest.json`** `data` lists for an already-installed module. **`all`** expands to every module with **`state = installed`** in **`sys.module`** (combined with **`sales,all`** dedupes **sales**).

---

### Build a release binary

| Command | Use case |
| ------- | -------- |
| `make build` | Writes **`sumeru`** in the repo root from `./cmd/sumeru` |
| `go build -o sumeru ./cmd/sumeru` | Equivalent manual build |
| `./sumeru -c sumeru.conf` | Run the compiled binary (no `go run` overhead) |

---

### Makefile targets

| Target | Command | Use case |
| ------ | ------- | -------- |
| `make generate` | `go generate ./cmd/sumeru` | Refresh **`cmd/sumeru/zimports.go`** from **`addons_path`** (see **`sumeru-import-gen`** flags in README) |
| `make run` | `go run ./cmd/sumeru -- -c sumeru.conf` | Quick dev server |
| `make run EXTRA_RUN_FLAGS='…'` | Appends flags (e.g. **`-p 9090`**, **`-d mydb`**, **`-u sales`**) | Same as a one-line **`go run`** with extra CLI flags |
| `make build` | `go build -o sumeru ./cmd/sumeru` | Produce **`sumeru`** binary |
| `make css` | Prints a reminder | There is **no** Sass build; edit **`core/engine/assets/css/*.css`** directly |
| `make help` | Lists targets | Discover available `make` commands |

---

### CLI flags (all entrypoints)

These apply to **`go run ./cmd/sumeru --`**, **`./sumeru.sh`**, and **`./sumeru`**:

| Flag | Use case |
| ---- | -------- |
| `-c <file>` | Path to INI config (default in script: **`sumeru.conf`** or **`SUMERU_CONF`**) |
| `-d <name>` | PostgreSQL **`dbname`** for this run; overrides INI **`db_name`** |
| `--database <name>` | Same as **`-d`** if set (**`--database`** wins when both are passed) |
| `--http-port <port>` | HTTP listen port; overrides INI **`http_port`** |
| `-p <port>` | Shorthand for **`--http-port`** (**`-p`** wins when both are passed) |
| `-i mod1,mod2` | **Install** listed modules (comma-separated); one or many |
| `-u mod1,mod2` | **Update** listed modules from disk; **`-u all`** updates every installed module |
| `--stop-after-init` | After **`-i`** / **`-u`**, exit **without** starting the HTTP server |

---

### Dependencies and housekeeping

| Command | Use case |
| ------- | -------- |
| `go mod download` | Fetch modules (CI, air-gapped prep) |
| `go build ./...` | Compile-check all packages under the module |
| `lsof -ti:8080 \| xargs kill -9` | Free port **8080** if the server is stuck (use the port you passed with **`-p`** / **`--http-port`** or INI **`http_port`**) |

---

## UI and static assets

Styles are **plain CSS** under `core/engine/assets/css/` (no Tailwind in markup; layout uses **`sum-*`** classes and design tokens). The global stack is **`DefaultStylesheetURLs()`** in `core/engine/assets/stylesheets.go`.

| File | Responsibility |
| ---- | ---------------- |
| `sumeru-theme.css` | **Branding only**: `:root` colors, typography, radii, shadows, layout tokens |
| `sumeru-base.css` | Document defaults, scrollbars, tabular number fonts, screen-reader helpers (e.g. `.sr-only`) |
| `sumeru-shell.css` | Top bar, sidebar, workspace grid, breadcrumbs, activity dock chrome |
| `sumeru-messages.css` | Activity panel **Messages** tab: thread, composer, empty state |
| `sumeru-views.css` | Form sheets, list tables, notebooks, read/write field chrome |
| `sumeru-compat.css` | Legacy **`.field`** blocks for login/setup templates |
| `sumeru-ai.css` | Optional AI shell widget styles (used when **`sumeru_ai`** is loaded) |
| `sumeru-login.css` | Standalone login card |
| `sumeru-pages.css` | Standalone pages (e.g. app logs table) |
| `sumeru-apps.css` | **`/web/apps`** catalog (linked per-page via `ViewStylesheetURLs`) |
| `sumeru-home.css` | **`/web/home`** app hub (per-page stylesheet) |
| `sumeru-settings-hub.css` | **`/web/settings`** overview (per-page stylesheet) |
| `sumeru-workspace.css` | **`/web`** workspace extras (tree, forms, toolbars) |

Per-addon optional **`static/css/theme-overrides.css`** is served as **`/static/addon-css/<module>.css`** after the core list. Optional **`brand_css`** in config loads after those. **`logo_path`** drives **`/static/app-logo`** in the header.

### HTML render package (`core/engine/render/`)

Workspace HTML is built from XML views in small, single-purpose Go files (same **`package render`**, unchanged public API):

| File | Role |
| ---- | ---- |
| `render_types.go` | `PageData`, `ViewRecordData`, hook registries (`RegisterShellHook`, `RegisterNotebookHook`) |
| `render_helpers.go` | Shared helpers (`recStr`, `rowOpenURL`, `formFieldReadonly`, …) |
| `workspace_chrome.go` | Edit / Save / Cancel bar and record `<form>` wrapper |
| `view_render.go` | **`RenderView`** — dispatches by view type, builds shell `PageData`, chatter |
| `form_render.go` | **`RenderForm`** and form sheet/header/notebook/field markup |
| `activity_chatter_render.go` | Activity panel message thread + composer |
| `user_security_render.go` | User form **Access rights** notebook tab |
| `tree_render.go`, `kanban_render.go`, `pivot_render.go` | List / kanban / pivot HTML |

Templates live under `core/engine/templates/` (`base.html`, inner pages). Client script entry is `core/engine/assets/js/core.js` (sidebar, activity panel, forms, messages composer, etc.).

---

## Project layout (short)

| Path | Use case |
| ---- | -------- |
| `core/orm/` | PostgreSQL-backed models, CRUD, registry |
| `core/engine/` | Module XML (`<sumeru>`, `<data>`, views), view inherit, HTML render, **`templates/`**, **`assets/`** |
| `core/server/` | INI **`config/`**, **`run.go`**, HTTP **`web/`** handlers |
| `core/module/` | Addon discovery, install/update, XML sync to DB |
| `cmd/sumeru/` | Default server **`main`** + generated **`zimports.go`** (blank imports for addon registration) |
| `core/base/` | **Stable Go API** for addons: config, DB bootstrap, model registration, queries (struct inputs). Default modules: **`core/base/user/`**, **`core/base/company/`**. |
| `core/base/base/` | Core **`base`** filesystem module (manifest only; always under prepended **`core/base`**) |
| `core/module/addon_template/` | Authoring reference for new addons (`manifest.template.json`, `MODULE_STANDARD.txt`) |
| `addons/` | Extra installable apps (`sales`, `crm`, `inventory`, …); **`user`** and **`company`** ship under **`core/base/`** |
| `sumeru.conf` | Runtime INI |
| `sumeru.sh` | Bash wrapper forwarding all flags |
| `Makefile` | `run`, `build`, `css`, `help`, `generate` |
| `../sumeru_custom_addons/` (monorepo) | Optional thin runner + INI + generated **`addonimports/`**; see section *Standard `sumeru` + workspace `sumeru_custom_addons`* above |
