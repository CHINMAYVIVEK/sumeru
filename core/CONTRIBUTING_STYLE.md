# Sumeru core code style

Conventions for `sumeru/core`. Reference implementation: [`orm/schema_migrate.go`](orm/schema_migrate.go).

## File layout

- **File header** — what the file owns and how it relates to siblings (2–4 lines).
- **Section breaks** — `// --- Topic ---` when one file covers several concerns.
- **Package docs** — each major package has a `doc.go` with architecture notes.

## Naming

| Area | Convention |
|------|------------|
| HTTP handlers | `(w http.ResponseWriter, r *http.Request)` |
| Context | always `ctx`; never name a parameter `context` |
| Exported API | verbs: `Run*`, `Ensure*`, `Load*`, `Build*`, `Find*` |
| Unexported helpers | describe the job: `backfillSysMenuModule`, `widenColumnToText` |
| Handler locals | full words: `viewRecord`, `actionData`, `menuID` — avoid `vr`, `vd` in large functions |
| Request structs | use named types when passing 4+ related values (`WorkspaceRequest`, `ViewLoadResult`) |

## Comments

- Explain **why**, not what the next line does.
- Skip godoc that repeats the function name.
- Avoid numbered step lists in source; use named functions instead.
- Document non-obvious business rules and legacy upgrade paths.

## HTML in Go (`engine/render`)

- User-facing strings: always `template.HTMLEscapeString` (or `html/template` partials).
- Repeated field chrome: shared helpers in `html_helpers.go`.
- Partial templates: only when the same block appears 3+ times.
- Inline SVG: extract to constants or `writeIcon` when duplicated.

## Deprecated APIs

When touching label/title code, replace `strings.Title` with `golang.org/x/text/cases`.

## Refactor discipline

- One behavioral change per PR when possible; style-only diffs are fine in dedicated passes.
- Run `go test ./...` from `sumeru/` before pushing.
- Internal renames stay inside `core/`; addon-facing APIs go through `core/sdk`.
