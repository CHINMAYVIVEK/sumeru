# Naming Conventions

Sumeru uses a strict naming convention for models to ensure architectural clarity and prevent namespace collisions. This system is inspired by the need to separate framework metadata from foundational business entities.

> **Developer hub:** [Documentation index](index.md) · [Models and fields](models_fields.md)

## Namespaces

### 1. `sys.*` (System / Framework)
The `sys.` namespace is reserved for **System Metadata**. These models define the framework's behavior and UI structure.
- **Rules**: If a model is required for the engine to render or secure data, it belongs in `sys.`.
- **Examples**:
    - `sys.model`: Registry of all available models.
    - `sys.field`: Definitions of model fields.
    - `sys.view`: XML view definitions (Form, Tree, Kanban).
    - `sys.rule`: Row-level security definitions.

### 2. `core.*` (Core / Foundational)
The `core.` namespace is for **Global Business Resources**. These are the primary entities that exist across all ERP installations regardless of installed modules.
- **Rules**: If a model represents a real-world entity used by multiple business apps, it belongs in `core.`.
- **Examples**:
    - `core.user`: System users.
    - `core.partner`: Contact entities (Customers, Vendors, Employees).
    - `core.company`: Legal organizational entities.
    - `core.currency`: Global financial data.

### 3. Module Namespaces
Custom modules should use their own short prefix to avoid collisions.
- **Example**: `crm.lead`, `sales.order`, `inventory.stock`.

## Why Not Dots?
While we use dots in the *string identifier* (e.g., `"core.user"`), the underlying database tables follow the snake_case convention with underscores (e.g., `core_user`).

## Go packages: `sdk` vs addon `base`

| Path | Go package | Role |
| ---- | ---------- | ---- |
| `sumeru/core/sdk` | `sdk` | Stable facade for addon/integration code (`RegisterModel`, field types, Search/Create, …). Prefer this over importing `sumeru/core/orm` directly. |
| `sumeru/addons/base` | `base` | Installable kernel addon (manifest name `"base"`). Models live under `addons/base/models` (`package models`). |

Do **not** rename the addon package or technical module name `"base"`. Only the former Go facade `sumeru/core/base` is now `sumeru/core/sdk`.

HTTP routes register through **`sumeru/core/server/router`** (`router.Register`, …).