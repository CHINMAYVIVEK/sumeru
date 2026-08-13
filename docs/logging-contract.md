# Sumeru logging contract

All framework components emit structured JSON logs via `sumeru/core/applog` (`Info`, `Warn`, `Debug`, `Error` with an `Event` struct). Do not import stdlib `log` in server or addon code.

**Pre-init:** Before `SetupFromConfig`, use `BootstrapFatal` (plain stderr + exit). After setup, use `Fatal` with `component=server`.

**Convenience helpers:** `InfoMsg`, `WarnMsg`, and `DebugMsg` wrap `Event` with `Status` set for readable call sites.

## Canonical shape

Top-level fields are minimal and universal. Module-specific data lives in **`context`**.

```json
{
  "time": "2026-08-13T11:12:08.123Z",
  "level": "INFO",
  "message": "Record created successfully",
  "request_id": "req_01JXYZ",
  "component": "orm",
  "module": "account",
  "operation": "create",
  "status": "success",
  "duration_ms": 18,
  "context": {
    "resource": "account.move",
    "resource_id": 18293,
    "user_id": 42,
    "company_id": 7
  }
}
```

| Field | Required | Notes |
|-------|----------|-------|
| `time` | yes | From slog JSON handler only — do not add `log_ts` |
| `level` | yes | INFO / WARN / DEBUG / ERROR |
| `message` | yes | Human-readable, e.g. `"Record created successfully"` |
| `request_id` | when available | HTTP middleware / RPC |
| `component` | yes | `orm`, `module`, `web`, `rpc`, `scheduler`, `event`, `server`, `render`, `applog` |
| `module` | optional | Declaring addon (`account`, `base`, …) |
| `operation` | yes | `create`, `update`, `delete`, `search`, `install`, … |
| `status` | yes | `success`, `failure`, `partial` |
| `duration_ms` | optional | Operation wall time |
| `context` | optional | Nested map: `resource`, `resource_id`, `user_id`, `company_id`, errors |

On failure, set `status=failure` and include `context.error`.

## Framework roots (auto-logging)

| Layer | Entry point | What is logged |
|-------|-------------|----------------|
| ORM | `logORMOperation` / `mutateRecord` | Create/Update/Delete at INFO; Search/SearchOne at DEBUG |
| Module sync | `syncWarn` / `linkXMLRecord` | XML upsert warnings |
| Web | `WebLogEvent`, `WebLogNavigation`, `renderShellPage` | Handler errors; navigation audit at INFO |
| RPC | `rpc_json.go` | model/method/status |
| HTTP | `SecurityMiddleware` | Request lifecycle (Debug; visible when `dev_mode=true`) |

### Web navigation audit (INFO)

Successful UI navigation emits one structured line via `WebLogNavigation`:

| `operation` | Handler | Context fields |
|-------------|---------|----------------|
| `view_open` | `WebHandler` | `menu_id`, `action_id`, `model`, `view_type`, `record_id`, `edit` |
| `module_hub` | `HomeDashboardHandler` | `menu_id`, `tile_count`, `search` |
| `apps_open` | `AppsHandler` | `layout`, `filter`, `scope`, `search`, optional `module` |
| `company_switch` | `SwitchCompanyPost` | `company_id`, `user_id` |
| `module_action` | `ModuleActionHandler` | `do`, `module` |

Local development: set `dev_mode=true` in `sumeru.conf` to also see DEBUG logs for HTTP requests, ORM reads, and view renders on stdout.

New addons inherit logging by calling ORM/RPC APIs — no extra logging code required for CRUD.

## Usage

```go
applog.Info(ctx, applog.Event{
    Message:   "Record created successfully",
    Component: "orm",
    Module:    "account",
    Operation: "create",
    Status:    "success",
    Duration:  time.Since(start),
    Context: map[string]interface{}{
        "resource":    "account.move",
        "resource_id": id,
    },
})
```

Ingestion, storage, indexing, and distributed tracing are follow-ups; this contract is stable for local JSON logs today.
