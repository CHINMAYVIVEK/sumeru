# Developer documentation

Task-oriented guides for people extending **Sumeru** (addons, views, security, and the web UI). Concepts live here; step-by-step recipes are under [**How-tos**](../howtos/index.md).

## Where to start

| You want to… | Read |
| ------------ | ---- |
| Understand the stack at a glance | [Architecture](architecture.md) |
| Add or change a module | [Addons overview](addons_overview.md) |
| Define tables and fields in Go | [Models and fields](models_fields.md) |
| XML views, actions, menus | [Views, actions, and menus](views_and_actions.md) |
| Groups, ACLs, record rules | [Security and access](security_access.md) |
| HTML shell, hooks, CSS | [Web layer](web_layer.md) |
| Available menu icons | [Web Icons](icons.md) |
| Naming `sys.*` vs `core.*` | [Naming conventions](naming_conventions.md) |
| CLI, code gen, scaffolding | [Tooling](tooling.md) |

## How-tos (recipes)

See **[`docs/howtos/`](../howtos/index.md)** for copy-paste workflows: scaffold an addon, install/update modules, wire menus, run from `sumeru_custom_addons`, and common pitfalls.

## Reference stubs (RST)

Long-form reference placeholders live under [`docs/refrence/`](../refrence/) (ORM, views, security, translation). Prefer the Markdown developer guides above for day-to-day work until those RST pages grow.

## Related repo docs

- **[`sumeru/README.md`](../../README.md)** — `sumeru.conf`, CLI flags, CSS and render package tables.
- **[`README.md`](../../../README.md)** (monorepo root) — quick start and doc index.
