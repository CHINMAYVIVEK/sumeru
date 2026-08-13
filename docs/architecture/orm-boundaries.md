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

Sumeru uses a **two-layer rule**:

1. **Identifiers** (tables, columns, ORDER BY fields) — validated and quoted via `QuotedTableForModel`, `QuotedColumnForModel`, `QuotedPermColumnForOp`, or compile-time `MustQuotedTableName`.
2. **Values** (user input, domain literals, limits) — always bound with `$1`, `$2`, … placeholders. Never interpolate request strings into SQL text.

Logical **model names** are lowercase segments joined only by `.` (e.g. `core.user`, `sys.config.parameter`). No underscores, whitespace, or other symbols in the logical name. Physical tables replace `.` with `_` after validation (`ModelToTableName` / `QuotedTableName`).

Every dynamic SQL path must:

1. Resolve tables via `QuotedTableForModel` / `QuotedTableName` (registered + validated).
2. Resolve columns via `QuotedColumnForModel` (whitelist against `model.Fields()` + `id`) or `QuotedPermColumnForOp` for ACL permission columns.
3. Bind all **values** with `$n` placeholders — never interpolate user strings into SQL.

**Forbidden:** building WHERE/ORDER BY/LIMIT from unchecked strings; splicing model or field names from HTTP/RPC/domain without validation.

**Allowed:** `fmt.Sprintf("SELECT * FROM %s WHERE %s", quotedTable, whereClause)` when `quotedTable` comes from `Quoted*` helpers and `whereClause` uses only `$n` placeholders.

Run `make check-sql` (see [`scripts/check_sql_safety.sh`](../scripts/check_sql_safety.sh)) before committing raw SQL changes.

## Side effects

Business mutation, audit, and outbox event rows commit together via `runMutationTx` / `execute*Mutation` in `crud_mutate.go`. All Create, Update, and Unlink paths enter this pipeline — no autocommit business-row mutations outside it (except documented bootstrap/schema bypass). Publishing to the in-process bus happens after commit (or via outbox worker). Never emit `record.*` for data that did not commit.

### Intentional stubs (not dead code)

| Feature | Status |
|---------|--------|
| **Outbox drain** | Rows are enqueued on CRUD; no publisher worker yet (`sys_outbox.go` TODO). |
| **Automation server actions** | Event subscriber logs matches only; execution is not wired. |
| **Cron `code` field** | Scheduled jobs log the field; never `eval` — use registered handlers only. |
| **Pivot view** | Placeholder renderer; use tree/kanban/form until implemented. |

Outbox rows are enqueued today; a drain worker is a follow-up. Side-effect failures are logged with `status=partial` via `applog.Warn`.

## Structured logging

ORM mutations, module sync, web handlers, and RPC emit JSON via `sumeru/core/applog` (`Event` API). See [`docs/logging-contract.md`](../logging-contract.md) for the canonical field schema (`message`, `component`, `operation`, `status`, nested `context`).

## Globals / Runtime

Package-level `orm.DB` and `orm.Registry` remain for bootstrap compatibility. New code should prefer `runtime.Runtime` (DB, registry wrapper, security cache, events, modules). Do not add new package-level singletons without documenting why.
