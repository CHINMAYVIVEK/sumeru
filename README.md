# Sumeru

**Sumeru** is an experimental ERP-style web application in Go (this repository’s Go module is **`sumeru`**): PostgreSQL-backed ORM, **addons** (module XML under `<sumeru>` + manifests), model sync on startup, and a web UI (XML views, plain CSS, shell with sidebar and activity panel).

- **Entry binary:** `cmd/sumeru/main.go` → **`sumeru/core/server`** (`server.Run`). Library code under **`core/`** has no `main`; addon code uses **`sumeru/core/sdk`** as the stable API.
- **Addon Go code:** import **`sumeru/core/sdk`** for models (stable API); avoid importing **`sumeru/core/orm`** directly in new code.

**Documentation:** **[`docs/developer/index.md`](docs/developer/index.md)** (concepts and reference layout) · **[`docs/howtos/index.md`](docs/howtos/index.md)** (step-by-step recipes) · user intro in **`docs/user/`** · RST stubs in **`docs/refrence/`**.

**Convention:** path-related INI values resolve from the **INI file’s directory** (and optional **`sumeru_home`** for default assets/templates). When running only from the **`sumeru`** repo root with **`sumeru.conf`** there, that matches the old “cwd = repo root” habit; prefer absolute paths in production.

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
| 2. Configure | Copy **`sumeru.conf.example`** → **`sumeru.conf`** and edit | Point `db_*`, `http_port`, `addons_path` at your environment |
| 3. Download modules | `go mod download` | Populate the Go module cache (CI or fresh clone) |
| 4. Run tests | `go test ./...` from the **`sumeru/`** module root | Unit and smoke tests live under **`test/`** (`sumeru/test/...`) |

---

## Standard `sumeru` + `sumeru_addons` + workspace `sumeru_custom_addons`

Keep **`sumeru`** and **`sumeru_addons`** pull-only for most teams. **Run the HTTP server from `sumeru_custom_addons`** so one INI can list core addons (`../sumeru/addons`), standard business addons (`../sumeru_addons`), and your own `./addons` without editing the standard trees.

| Piece | Role |
| ----- | ---- |
| **`sumeru/`** | Core engine + **`addons/`** (`base`, `sumeru_ai`, …). `go generate ./cmd/sumeru` reads **`sumeru.conf.example`** and refreshes **`cmd/sumeru/zimports.go`** (core imports only). |
| **`sumeru_addons/`** | Separate `go.mod` (`module sumeru_addons`) for standard business apps; depends only on **`sumeru`**. |
| **`sumeru_custom_addons/`** | Your workspace: `replace` to **`../sumeru`** and **`../sumeru_addons`**, **`make generate`** → **`addonimports/zimports.go`**, **`make run`**. |

From **`sumeru_custom_addons`**: copy **`sumeru.conf.example`** to **`sumeru.conf`**, then **`make run`**. Details: **[`../sumeru_custom_addons/README.md`](../sumeru_custom_addons/README.md)**.

**`sumeru-import-gen`** flags (also used by `go generate ./cmd/sumeru`):

| Flag | Default | Use case |
| ---- | ------- | -------- |
| **`-root`** | cwd walk-up | Standard **`sumeru`** repo root (directory with **`go.mod`**). |
| **`-config`** | `sumeru.conf.example` (via `go generate` in **`cmd/sumeru`**) | INI path; absolute, or relative to **`-root`**. |
| **`-out`** | `cmd/sumeru/zimports.go` | Output file; if **relative**, resolved against **`-root`** (use an **absolute** path to write under **`sumeru_custom_addons`**). |
| **`-package`** | `main` | Generated file’s package name (e.g. **`addonimports`** for the workspace). |

**Custom Go addons** in **`sumeru_custom_addons/addons/`** ship as **`sumeru_custom_addons/addons/<name>`** in generated imports; addons under **`sumeru/addons/`** use **`sumeru/addons/<name>`**. Standard business addons under **`sumeru_addons/`** use **`sumeru_addons/<name>`** when listed on **`addons_path`** from the custom workspace.

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
| `addons_path` | Comma-separated addon roots (e.g. **`addons`** for kernel apps including **`base`**). Later roots **override** duplicate module **names**. Relative segments are resolved from the **INI file’s directory**. |
| `sumeru_home` | Optional. Directory of the standard **`sumeru`** checkout. Relative values are resolved from the **INI file’s directory**. When set, default **`assets_path`** / **`templates_path`** (if omitted in the INI) are under this tree. Use for configs outside the repo (e.g. **`sumeru_custom_addons/sumeru.conf`**). |
| `assets_path` | Static files root (default **`core/engine/assets`**; resolved from INI dir unless absolute; see **`sumeru_home`**) |
| `templates_path` | HTML templates (default **`core/engine/templates`**; same resolution rules) |
| `logo_path` | Optional image file; served at **`/static/app-logo`** (`.svg`, `.png`, `.jpg`/`.jpeg`, `.webp`) |
| `company_display_name` | Optional header chip; if empty and **`company`** module is installed, first **`core.company`** name is used |
| `user_display_name` | Optional header label; if empty and **`user`** module is installed, first **`core.user`** display is used |
| `brand_css` | Optional extra CSS file; linked as **`/static/brand.css`** after view stylesheets |
| `log_enabled` | Default **true**. Set **false** to disable application JSON logging (Zap no-op; **`log`** package output discarded after init). |
| `log_timezone` | **`UTC`**, **`Local`**, or an IANA name (e.g. **`Asia/Kolkata`**) for Zap timestamps and the **`log_tz`** field on **`applog.L(ctx)`** lines. |
| `log_stdout` | Default **true** — JSON logs to **stdout** (Uber **Zap**). Set **false** for file-only sinks. |
| `log_file` | Optional log path (absolutized like other paths). When set with **`log_stdout`**, both receive the same JSON lines. |
| `log_rolling` | Default **false** — append-only file. Set **true** with **`log_file`** for **lumberjack** size rotation (typical VPS); keep **false** on Kubernetes and collect **stdout** only. |
| `log_max_size_mb` | Max megabytes per file before rotation (default **100** when **`log_rolling`** is true and this is **0**). |
| `log_max_backups` | Retained rotated files ( **0** = lumberjack default ). |
| `log_max_age_days` | Delete rotated files older than **N** days ( **0** = no age limit ). |
| `dev_mode` | Default **false** when omitted. When **true**, debug-level Zap logs and dev-only server behavior. Uses the same boolean parsing as **`log_stdout`** (`true` / `false` / `1` / `0` / `yes` / `no` / `on` / `off`). |

### HTTP API (Odoo-style JSON-RPC)

| Endpoint | Use case |
| -------- | -------- |
| **`GET /api/health`** | Liveness probe; returns **`{"ok":true}`** (no auth). |
| **`POST /api/rpc`** | JSON body **`{"model":"core.company","method":"search_read","args":[...],"kwargs":{}}`** with the same session cookie as the web UI. Methods: **`search`**, **`search_read`**, **`read`**, **`create`**, **`write`**, **`unlink`**. Optional wrapper: put the same keys inside **`params`** for Odoo-style envelopes. |

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
| `make generate` | `go generate ./cmd/sumeru` | Refresh **`cmd/sumeru/zimports.go`** from **`sumeru.conf.example`** (tracked; copy to **`sumeru.conf`** for **`make run`**) |
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

**Canonical reference:** keep the global CSS file table (below) only in this README. Other docs should link here instead of duplicating the list.

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
| `core/sdk/` | **Stable Go API** for addons: config, DB bootstrap, model registration, queries (struct inputs). |
| `core/server/router/` | Addon-extensible HTTP route registry |
| `core/module/addon_template/` | Authoring reference for new addons (`manifest.template.json`, `MODULE_STANDARD.txt`) |
| `addons/` | Core installable apps shipped with **`sumeru`** (`base`, `mail`, `automation`, `sumeru_ai`, …) |
| `sumeru.conf` | Local runtime INI (gitignored if you use a local policy); copy from **`sumeru.conf.example`**. |
| `sumeru.conf.example` | Tracked template + **import-gen** input for **`go generate ./cmd/sumeru`**. |
| `sumeru.sh` | Bash wrapper forwarding all flags |
| `Makefile` | `run`, `build`, `css`, `help`, `generate` |
| `../sumeru_addons/` (monorepo) | Standard business addons module; see sibling **`sumeru_addons/README.md`**. |
| `../sumeru_custom_addons/` (monorepo) | **Recommended** thin runner + INI + **`addonimports/`**; see section *Standard `sumeru` + `sumeru_addons` + workspace `sumeru_custom_addons`* above |
