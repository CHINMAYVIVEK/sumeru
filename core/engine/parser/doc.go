// Package parser converts addon XML and related files into structured Go values (views, records, menus).
//
// Dependency rules (keep this package easy to test and reuse):
//   - Do not import sumeru/core/orm, sumeru/core/server, or sumeru/core/engine/render.
//   - I/O is limited to reading files passed in by callers; no database or HTTP.
package parser
