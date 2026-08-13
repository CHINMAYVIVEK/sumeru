# ORM vs ERP Boundaries

Sumeru keeps a hard line between **ORM core** and **ERP behavior**. Crossing that line is how frameworks become unmaintainable.

## ORM core owns

- Model registry and field definitions
- CRUD (`Create`, `Update`, `Unlink`, `Search*`)
- Domains → SQL
- ACL (`sys.access`) and record rules (`sys.rule`)
- Field-level write ACL (`sys.field.access`)
- Transactions and row locking for secure mutations
- Audit hooks and outbox rows (side-effect contracts)
- Schema sync for declared fields

## ERP modules own (via composition)

- Workflow / stage approvals
- Company policies and multi-company UX
- Mail / chatter / activity panel
- Automation (event subscribers)
- Import pipelines and business validation

Use **interceptors**, **events**, and **services** — do not grow `core/orm` with business concepts.

## Odoo-compatible RPC vs Go-native internals

| Layer | Contract |
|-------|----------|
| HTTP `POST /api/rpc` | Odoo-inspired shapes (`model`, `method`, `args`, `kwargs`, optional `params`) |
| Internal Go API | Native: `PrepareValues`, `Update(ctx, model, domain, values)`, `SearchPage` |

Compatibility stops at the boundary. Internal design is Go-first, not “Odoo in Go”.

## Security is query-shaped

Authorization decisions must not sit in a check-then-mutate window:

1. Compile ACL + record rules into SQL predicates (and/or use `SELECT … FOR UPDATE` in the same transaction).
2. Mutate with `WHERE id = ? AND <predicate>`.
3. Treat zero rows affected as access denied / not found (no existence leak).

## Values pipeline (create / write)

```
request → model lookup → field whitelist → field ACL → type coercion → validation → record rules → SQL
```

`PrepareValues` is the shared sanitizer for Create, Update, and Upsert.

## SQL identifiers (injection safety)

Logical **model names** are lowercase segments joined only by `.` (e.g. `core.user`, `sys.config.parameter`). No underscores, whitespace, or other symbols in the logical name. Physical tables replace `.` with `_` after validation (`ModelToTableName` / `QuotedTableName`).

Every dynamic SQL path must:

1. Resolve tables via `QuotedTableForModel` / `QuotedTableName` (registered + validated).
2. Resolve columns via `QuotedColumnForModel` (whitelist against `model.Fields()` + `id`).
3. Bind all **values** with `$n` placeholders — never interpolate user strings into SQL.

## Side effects

Business mutation, audit, and outbox event rows commit together. Publishing to the in-process bus happens after commit (or via outbox worker). Never emit `record.*` for data that did not commit. Outbox rows are enqueued today; a drain worker is a follow-up.

## Globals / Runtime

Package-level `orm.DB` and `orm.Registry` remain for bootstrap compatibility. New code should prefer `runtime.Runtime` (DB, registry wrapper, security cache, events, modules). Do not add new package-level singletons without documenting why.
