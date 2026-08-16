// Package orm is Sumeru's data layer: model registry, CRUD, security, and schema.
//
// Layout:
//   - registry.go, sys_*.go — model definitions and registration
//   - crud_*.go — create, read, update, delete, search, domains
//   - security_*.go, record_rules.go, field_access.go — ABAC and ACL
//   - schema_sync.go — additive DDL and indexes from model Fields() on install/startup
//   - ui_view_lookup.go — resolve sys.view rows for the workspace UI
//
// Addons should use sumeru/core/sdk rather than importing this package directly.
package orm
