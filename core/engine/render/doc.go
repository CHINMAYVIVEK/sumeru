// Package render builds workspace HTML from parser views and ORM-backed record data.
//
// Dependency rules:
//   - Do not import sumeru/core/server/web (avoid render ↔ HTTP cycles). Pass URLs, flags, and
//     branding through PageData and related structs instead of reaching into handlers.
//   - sumeru/core/engine/parser is the source for view XML shapes.
//   - sumeru/core/orm is allowed for display-oriented reads that already exist in templates/helpers.
//   - Prefer keeping new configuration on PageData over new imports from sumeru/core/server/… .
package render
