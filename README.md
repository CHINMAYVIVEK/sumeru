# Sumeru

[![Go](https://img.shields.io/badge/Go-1.26.2+-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Pre-Alpha](https://img.shields.io/badge/Status-Pre--Alpha-critical?style=for-the-badge)](https://github.com/ProjectMeru/sumeru)

> [!CAUTION]
> ### 🚧 Pre-Alpha Software
>
> Sumeru is **pre-alpha software**. It is under active development and is **not ready for production or commercial use**.
>
> - **No production use.** Do not deploy to production or run live business workloads. Stability, security, and data integrity are not guaranteed.
> - **Not for sale.** Do not offer, resell, license, or deploy Sumeru to customers. This is not a commercial product.
> - **Evaluation only.** Use for local development, testing, and feedback at your own risk.
>
> APIs, data models, and behavior may change without notice. There is no migration guarantee and no production support.

**Sumeru** is an experimental ERP-style web application written in **Go**. It provides a PostgreSQL-backed ORM, installable **addons** (module XML under `<sumeru>` + manifests), model sync on startup, and a web UI (XML views, plain CSS, shell with sidebar and activity panel).

This repository is the **core engine** (`module sumeru`). Most teams keep it pull-only and run the server from **`sumeru_custom_addons`**.

## Features

- Modular **addons** with manifests, XML views/menus, and Go model registration
- PostgreSQL ORM with model sync on startup
- Web shell: apps catalog, home, settings, tree/form/kanban workspaces
- Odoo-style JSON-RPC at **`POST /api/rpc`** (plus **`GET /api/health`**)
- Stable addon API via **`sumeru/core/sdk`** (prefer over importing **`sumeru/core/orm`** directly)

## Architecture

Sumeru is split into three repositories so you can pull updates to the engine and standard apps without mixing in customer-specific code.

```text
sumeru_custom_addons  ──replace + make generate──►  sumeru (core)
         │                                              │
         └──replace + addons_path──────────────────────►│
         │                                              ▼
         └──make run────────────────────────────►  HTTP server
                ▲
                └── also loads  sumeru_addons (standard business apps)
```

| Repository                 | Role                                                                                    | Remote                                                |
| -------------------------- | --------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| **`sumeru`**               | Core engine + kernel addons (`base`, `mail`, `sumeru_ai`, …). Pull-only for most teams. | `git@github.com:ProjectMeru/sumeru.git`               |
| **`sumeru_addons`**        | Standard business apps (CRM, Sales, Inventory, …). Pull-only.                           | `git@github.com:ProjectMeru/sumeru_addons.git`        |
| **`sumeru_custom_addons`** | Your workspace: custom addons, local INI, generated imports, and the process you run.   | `git@github.com:ProjectMeru/sumeru_custom_addons.git` |

**Entry binary (this repo):** `cmd/sumeru/main.go` → `sumeru/core/server` (`server.Run`). Library code under `core/` has no `main`.

## Prerequisites

| Requirement                               | Notes                |
| ----------------------------------------- | -------------------- |
| [Go](https://go.dev/dl/) **1.26.2+**      | See `go.mod`         |
| [PostgreSQL](https://www.postgresql.org/) | Application database |

## Quick start (recommended)

Clone all three repos as siblings, configure the custom workspace, generate blank-imports, then run.

```bash
mkdir -p ~/sumeru_erp && cd ~/sumeru_erp
git clone git@github.com:ProjectMeru/sumeru.git
git clone git@github.com:ProjectMeru/sumeru_addons.git
git clone git@github.com:ProjectMeru/sumeru_custom_addons.git

# Create an empty database matching db_name in your INI, e.g.:
#   psql -c "CREATE DATABASE sumeru;"

cd sumeru_custom_addons
cp sumeru.conf.example sumeru.conf   # edit db_* , http_port, addons_path
make replace-sumeru
make replace-sumeru-addons
make generate                        # → addonimports/zimports.go
make run                             # generate + go run
```

Open **`http://localhost:<http_port>`** (default **8080**). `/` redirects to **`/web/apps`**.

### Day-to-day updates

```bash
cd ../sumeru && git pull
cd ../sumeru_addons && git pull
cd ../sumeru_custom_addons && make generate && make run
```

Full workspace details: sibling **[`sumeru_custom_addons/README.md`](https://github.com/ProjectMeru/sumeru_custom_addons/blob/main/README.md)**.

### Optional: core-only run

Useful when you only need kernel addons under `sumeru/addons/`:

```bash
cd sumeru
cp sumeru.conf.example sumeru.conf   # edit db_* ; addons_path = addons
make generate                        # refreshes cmd/sumeru/zimports.go
make run                             # or: go run ./cmd/sumeru -- -c sumeru.conf
```

Install first apps (example), then serve:

```bash
go run ./cmd/sumeru -- -c sumeru.conf -i company,user --stop-after-init
go run ./cmd/sumeru -- -c sumeru.conf
```

---

## Configuration (`sumeru.conf`)

INI format: `key = value` under **`[options]`**. Lines starting with `#` or `;` are comments. Path-related values resolve from the **INI file’s directory** (and optional **`sumeru_home`** for default assets/templates). Prefer absolute paths in production.

Copy **`sumeru.conf.example`** → **`sumeru.conf`**.

| Key                                         | Use case                                                                                                                        |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `db_host`, `db_port`                        | PostgreSQL host and port                                                                                                        |
| `db_user`, `db_password`                    | Database credentials                                                                                                            |
| `db_name`                                   | Database name (overridable with **`-d`** / **`--database`**)                                                                    |
| `db_sslmode`                                | PostgreSQL SSL mode (e.g. `disable` for local dev)                                                                              |
| `http_port`                                 | HTTP listen port (default **8080**; overridable with **`-p`** / **`--http-port`**)                                              |
| `addons_path`                               | Comma-separated addon roots. Later roots **override** duplicate module names. Relative segments resolve from the INI directory. |
| `sumeru_home`                               | Optional path to the standard **`sumeru`** checkout; default assets/templates when those keys are omitted                       |
| `assets_path`, `templates_path`             | Static files and HTML templates (defaults under `core/engine/…`)                                                                |
| `logo_path`                                 | Optional image; served at **`/static/app-logo`**                                                                                |
| `company_display_name`, `user_display_name` | Optional header labels                                                                                                          |
| `brand_css`                                 | Optional extra CSS; linked as **`/static/brand.css`**                                                                           |
| `dev_mode`                                  | Default **false**. When **true**, debug-level logs and dev-only behavior                                                        |

**Logging:** `log_enabled`, `log_timezone`, `log_stdout`, `log_file`, `log_rolling`, and related rotation keys are documented in **`sumeru.conf.example`**.

### HTTP API

| Endpoint              | Use case                                                                                                                                                                             |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **`GET /api/health`** | Liveness; `{"ok":true}` (no auth)                                                                                                                                                    |
| **`POST /api/rpc`**   | Model RPC with session cookie or API key (see below)                                                                                                                                 |

#### `POST /api/rpc`

**Authentication:** session cookie (`sumeru_session`) from `/web/login`, or `X-API-Key: sk_…` / `Authorization: Bearer sk_…`.

**Request** (`Content-Type: application/json`):

Sumeru flat shape:

```json
{
  "model": "core.user",
  "method": "search_read",
  "args": [[["active", "=", true]], ["id", "login"]],
  "kwargs": { "limit": 50, "offset": 0 }
}
```

Odoo-style wrapper (also supported):

```json
{
  "params": {
    "model": "core.user",
    "method": "search",
    "args": [[]],
    "kwargs": {}
  }
}
```

**Response envelope** (always JSON; HTTP status reflects success or error class):

```json
{ "ok": true, "result": <method-specific>, "error": null }
```

```json
{
  "ok": false,
  "result": null,
  "error": {
    "code": "ACCESS_DENIED",
    "message": "access denied on core.user for operation read",
    "details": {}
  }
}
```

Legacy clients may ignore `ok` and continue checking `error == null`.

**Public methods**

| Method        | `args`             | `kwargs`          | `result`                         |
| ------------- | ------------------ | ----------------- | -------------------------------- |
| `search`      | `[domain?]`        | `limit`, `offset` | `[{record}, …]`                  |
| `search_read` | `[domain, fields]` | `limit`, `offset` | `[{record}, …]` (field projection) |
| `read`        | `[ids, fields?]`   | —                 | `[{record}, …]`; missing ids → `NOT_FOUND` with `details.missing_ids` |
| `create`      | `[values]`         | —                 | `int` (new id)                   |
| `write`       | `[ids, values]`    | —                 | `true`                           |
| `unlink`      | `[ids]`            | —                 | `true`                           |

Default `kwargs.limit` is **500**. `offset` skips rows after the ORM fetch (in-memory slice).

**Error codes** (representative HTTP status)

| `error.code`             | HTTP | Typical cause                          |
| ------------------------ | ---- | -------------------------------------- |
| `INVALID_JSON`           | 400  | Malformed JSON, empty body             |
| `INVALID_ARGS`           | 400  | Bad `args`/`kwargs` shape or arity     |
| `VALIDATION_ERROR`       | 400  | Missing `model` or `method`            |
| `INVALID_BODY`           | 400  | Request body could not be read         |
| `UNSUPPORTED_MEDIA_TYPE` | 415  | `Content-Type` is not JSON             |
| `PAYLOAD_TOO_LARGE`      | 413  | Body exceeds 4 MiB                     |
| `UNAUTHORIZED`           | 401  | No session or API key                  |
| `METHOD_NOT_ALLOWED`     | 403/405 | Unknown RPC method / wrong HTTP verb |
| `MODEL_NOT_FOUND`        | 404  | Model not in registry                  |
| `NOT_FOUND`              | 404  | Record id(s) not found on `read`       |
| `ACCESS_DENIED`          | 403  | ORM security rule                      |
| `INTERNAL_ERROR`         | 500  | Unexpected server failure              |

---

## CLI

All entrypoints (`go run ./cmd/sumeru --`, `./sumeru.sh`, `./sumeru`, and the custom-workspace `go run . --`) accept:

| Flag                               | Use case                                             |
| ---------------------------------- | ---------------------------------------------------- |
| `-c <file>`                        | INI config path                                      |
| `-d <name>` / `--database <name>`  | Override `db_name` (`--database` wins if both set)   |
| `-p <port>` / `--http-port <port>` | Override `http_port` (`-p` wins if both set)         |
| `-i mod1,mod2`                     | **Install** listed modules                           |
| `-u mod1,mod2` or `-u all`         | **Update** from disk; `all` = every installed module |
| `--stop-after-init`                | After `-i` / `-u`, exit without starting HTTP        |

Examples:

```bash
go run ./cmd/sumeru -- -c sumeru.conf -p 9090
go run ./cmd/sumeru -- -c sumeru.conf -i sales,crm --stop-after-init
go run ./cmd/sumeru -- -c sumeru.conf -u all -p 9090 --stop-after-init
make run EXTRA_RUN_FLAGS='-u sales -p 9090'
```

Use **`-u`** after changing `views/*.xml`, `menus.xml`, or `manifest.json` data lists for an already-installed module.

---

## Makefile (this repo)

| Target                   | Use case                                                                                 |
| ------------------------ | ---------------------------------------------------------------------------------------- |
| `make generate`          | `go generate ./cmd/sumeru`: refresh `cmd/sumeru/zimports.go` from `sumeru.conf.example` |
| `make run`               | Generate, then `go run ./cmd/sumeru -- -c sumeru.conf` (optional `EXTRA_RUN_FLAGS`)      |
| `make build`             | Generate, then `go build -o sumeru ./cmd/sumeru`                                         |
| `make bp NAME=my_module` | Scaffold a core-tree addon (`WITH_MODELS=1` optional)                                    |
| `make css`               | Reminder: plain CSS under `core/engine/assets/css/` (no Sass build)                      |
| `make help`              | List targets                                                                             |

In **`sumeru_custom_addons`**, use that repo’s Makefile (`make generate`, `make run`, `make replace-sumeru`, …) so imports are written under `addonimports/`, not into this tree.

**`sumeru-import-gen`** (used by `go generate` / workspace `make generate`):

| Flag       | Default                                                   | Use case                                                        |
| ---------- | --------------------------------------------------------- | --------------------------------------------------------------- |
| `-root`    | cwd walk-up                                               | Standard `sumeru` repo root                                     |
| `-config`  | `sumeru.conf.example` (via `go generate` in `cmd/sumeru`) | INI path                                                        |
| `-out`     | `cmd/sumeru/zimports.go`                                  | Output file (absolute path when writing under custom workspace) |
| `-package` | `main`                                                    | e.g. `addonimports` for the workspace                           |

---

## Project layout

| Path                          | Use case                                                       |
| ----------------------------- | -------------------------------------------------------------- |
| `core/orm/`                   | PostgreSQL-backed models, CRUD, registry                       |
| `core/engine/`                | Module XML, view inherit, HTML render, `templates/`, `assets/` |
| `core/server/`                | INI config, `run.go`, HTTP web handlers                        |
| `core/module/`                | Addon discovery, install/update, XML sync                      |
| `core/sdk/`                   | **Stable Go API** for addons                                   |
| `core/server/router/`         | Addon-extensible HTTP route registry                           |
| `core/module/addon_template/` | Authoring reference for new addons                             |
| `cmd/sumeru/`                 | Server `main` + generated `zimports.go`                        |
| `cmd/sumeru-import-gen/`      | Blank-import generator                                         |
| `cmd/sumeru-bp/`              | Addon scaffold tool                                            |
| `addons/`                     | Kernel apps shipped with this repo                             |
| `test/`                       | Unit and smoke tests (`go test ./...`)                         |
| `sumeru.conf.example`         | Tracked template + import-gen input                            |
| `sumeru.sh`                   | Bash wrapper forwarding CLI flags                              |
| `Makefile`                    | `generate`, `run`, `build`, `bp`, `help`                       |

---

## UI and static assets

Styles are **plain CSS** under `core/engine/assets/css/` (layout uses `sum-*` classes and design tokens). Global stack: `DefaultStylesheetURLs()` in `core/engine/assets/stylesheets.go`.

**Canonical reference:** keep the global CSS file table below only in this README. Other docs should link here instead of duplicating the list.

| File                      | Responsibility                                               |
| ------------------------- | ------------------------------------------------------------ |
| `sumeru-theme.css`        | Branding tokens (`:root` colors, typography, radii, shadows) |
| `sumeru-base.css`         | Document defaults, scrollbars, `.sr-only`, etc.              |
| `sumeru-shell.css`        | Top bar, sidebar, workspace grid, activity dock chrome       |
| `sumeru-messages.css`     | Activity panel Messages tab                                  |
| `sumeru-views.css`        | Form sheets, list tables, notebooks, field chrome            |
| `sumeru-compat.css`       | Legacy `.field` blocks for login/setup templates             |
| `sumeru-ai.css`           | Optional AI shell widget (`sumeru_ai`)                       |
| `sumeru-login.css`        | Standalone login card                                        |
| `sumeru-pages.css`        | Standalone pages (e.g. app logs)                             |
| `sumeru-apps.css`         | `/web/apps` catalog (per-page)                               |
| `sumeru-home.css`         | `/web/home` (per-page)                                       |
| `sumeru-settings-hub.css` | `/web/settings` (per-page)                                   |
| `sumeru-workspace.css`    | `/web` workspace extras                                      |

Per-addon optional `static/css/theme-overrides.css` is served as `/static/addon-css/<module>.css`. Optional `brand_css` loads after those. Workspace HTML lives under `core/engine/render/`; templates under `core/engine/templates/`; client entry `core/engine/assets/js/core.js`.

---

## Documentation

| Resource                                                                                                    | Contents                                                                                                                       |
| ----------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| This README                                                                                                 | Setup, config, CLI, Makefile, CSS                                                                                              |
| [`sumeru_addons/README.md`](https://github.com/ProjectMeru/sumeru_addons/blob/main/README.md)               | Standard business addon module                                                                                                 |
| [`sumeru_custom_addons/README.md`](https://github.com/ProjectMeru/sumeru_custom_addons/blob/main/README.md) | Workspace runner, `make generate`, custom addons                                                                               |
| Sibling `docs/` (local workspace)                                                                           | Developer guides and how-tos when checked out next to this repo (e.g. `../docs/developer/`). Not shipped inside this git tree |

---

## Contributing

See **[CONTRIBUTING.md](CONTRIBUTING.md)** for where to put changes, the generate/test loop, and PR expectations.

Please follow the **[Code of Conduct](CODE_OF_CONDUCT.md)**.

## Security

Report vulnerabilities privately. See **[SECURITY.md](SECURITY.md)**. Do not open public issues for undisclosed security problems.

## License

Licensed under the [Apache License, Version 2.0](LICENSE).
