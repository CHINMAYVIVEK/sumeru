# How to scaffold a new addon

**Goal:** create a valid module tree with manifest, Go model stub, starter views, menus, and security files.

## Prerequisites

- Go toolchain per **`sumeru/go.mod`**
- Repo root is the directory that contains **`sumeru/go.mod`**

## Steps

1. From **`sumeru/`** (or any cwd under the repo), run:

   ```bash
   go run ./cmd/sumeru-bp -- -bp my_module
   ```

   Use **`^[a-z][a-z0-9_]*$`** for **`my_module`** (e.g. **`sale_demo`**, not **`SaleDemo`**).

2. Regenerate **`cmd/sumeru/zimports.go`** so **`init()`** runs:

   ```bash
   make generate
   ```

3. Install into your dev database once:

   ```bash
   go run ./cmd/sumeru -- -c sumeru.conf -i my_module --stop-after-init
   ```

4. Start the server without **`--stop-after-init`** and open **Apps** to confirm the module appears; open your menu entry to reach the scaffolded action.

## Custom output directory

```bash
go run ./cmd/sumeru-bp -- -bp my_module -out addons
```

Default is already **`addons/`** under the detected repo root.

## See also

- [Install a module](install_module.md)
- [Tooling](../developer/tooling.md)
