# Contributing to Sumeru

Thanks for helping improve Sumeru. This guide covers where to put changes, how to develop locally, and what we expect in pull requests.

By participating, you agree to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Where to put your work

Use the right repository so core and standard apps stay pullable for everyone:

| Change type | Repository | Notes |
| ----------- | ---------- | ----- |
| Engine, ORM, server, web shell, kernel addons (`base`, `mail`, …) | **`sumeru`** (this repo) | Prefer `sumeru/core/sdk` from addon Go code; avoid new direct imports of `sumeru/core/orm` |
| Shared business apps (CRM, Sales, Inventory, …) | **`sumeru_addons`** | Depends only on `sumeru` |
| Customer-specific modules, branding, local runner | **`sumeru_custom_addons`** | Keep custom code under `addons/`; do not fork core for one-off features |

Most application teams **pull** `sumeru` and `sumeru_addons` and develop only in `sumeru_custom_addons`.

## Development setup

1. Clone the three siblings (see [README.md](README.md#quick-start-recommended)).
2. Create a PostgreSQL database and configure `sumeru_custom_addons/sumeru.conf`.
3. From `sumeru_custom_addons`:

   ```bash
   make replace-sumeru
   make replace-sumeru-addons
   make generate
   make run
   ```

4. After pulling core or standard addons:

   ```bash
   cd ../sumeru && git pull
   cd ../sumeru_addons && git pull
   cd ../sumeru_custom_addons && make generate
   ```

Core-only work can use `make generate` / `make run` from this repo (see README).

## Code generation

- **This repo:** `make generate` runs `go generate ./cmd/sumeru` and refreshes `cmd/sumeru/zimports.go` from `sumeru.conf.example`.
- **Custom workspace:** `make generate` writes `addonimports/zimports.go` from that workspace’s INI — do not generate custom imports into the `sumeru` tree.

Scaffold a new **core-tree** addon:

```bash
make bp NAME=my_module
# optional: make bp NAME=my_module WITH_MODELS=1
```

For custom modules, scaffold under `sumeru_custom_addons` (see that repo’s docs / `sumeru-bp` usage).

## Testing

From the `sumeru` module root:

```bash
go test ./...
go build ./...
```

Add or update tests under `test/` when you change ORM, server, or module behavior.

## Pull requests

- Keep diffs focused; one concern per PR when practical.
- Match existing naming and layout (`sys.*` / `core.*`, addon folder = technical name).
- Do **not** commit local secrets, `sumeru.conf` with real passwords, or credentials.
- Do not commit generated custom-workspace `addonimports/` unless that project explicitly tracks them.
- Describe *why* the change is needed and how you verified it (commands, ports, modules installed).
- For security-sensitive findings, follow [SECURITY.md](SECURITY.md) instead of a public PR discussion of exploits.

## Questions

Open a GitHub issue on [ProjectMeru/sumeru](https://github.com/ProjectMeru/sumeru) for design or bug discussion. For vulnerabilities, use the private process in [SECURITY.md](SECURITY.md).
