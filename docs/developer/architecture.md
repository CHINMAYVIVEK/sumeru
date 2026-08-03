# Architecture

Sumeru is a **Go** monolith with **PostgreSQL**, **XML-defined** data (views, menus, actions, security records), and a **plain CSS + server-rendered HTML** web UI.

## High-level flow

1. **`cmd/sumeru`** (or a thin wrapper in `sumeru_custom_addons`) calls **`sumeru/core/server`**, loads **`sumeru.conf`**, initializes the ORM, syncs schema, discovers addons from **`addons_path`**, then serves HTTP.
2. **`core/orm`** holds model metadata, SQL generation, CRUD, and access checks. Addon code should use **`sumeru/core/sdk`** instead of importing ORM internals directly.
3. **`core/module`** discovers directories with **`manifest.json`**, resolves **`depends`**, installs/updates modules, and loads **`manifest.data`** XML into the database when a module is installed.
4. **`core/engine`** parses addon XML (`<sumeru>`, `<data>`), resolves view inheritance, and **`core/engine/render`** turns views + records into HTML for **`/web`** routes.
5. Built-in kernel behavior for users/companies lives in the **`base`** addon under **`sumeru/addons/base/`** (Go models + XML).

## Main packages

| Path | Responsibility |
| ---- | ---------------- |
| `core/server` | Config, process lifecycle, HTTP routing (`web/`) |
| `core/server/router` | Addon-extensible HTTP route registry |
| `core/orm` | Models registry, migrations/sync, queries, ACL/rules |
| `core/sdk` | **Stable API** for addons (`RegisterModel`, `Search`, …) |
| `core/module` | Addon discovery, manifest, install/update, XML → DB |
| `core/engine/parser` | XML parsing for records, views, menus |
| `core/engine/render` | Workspace HTML, hooks, chatter |
| `addons/<name>/` | Installable apps (manifest + `models/` + `views/` + optional `security/`) |

## Web routes (mental model)

| URL | Role |
| --- | ---- |
| `/web/apps` | Module catalog (install/update) |
| `/web/home` | App hub for signed-in users |
| `/web/settings` | Settings overview and shortcuts |
| `/web?…` | Workspace: tree / form / kanban for a model |

See **`sumeru/README.md`** for stylesheet and render file layout.
