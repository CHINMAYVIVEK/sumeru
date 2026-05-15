# How to install a module

**Goal:** register a discovered addon in **`sys.module`** and load its **`manifest.data`** XML.

## Steps

1. Ensure the module directory sits under a path listed in **`addons_path`** in **`sumeru.conf`** (comma-separated). Later entries override same **`manifest.name`**.

2. If the addon has Go **`init()`** registration, ensure **`cmd/sumeru/zimports.go`** blank-imports the addon root package:

   ```bash
   make generate
   ```

3. Install (one-time per database):

   ```bash
   go run ./cmd/sumeru -- -c sumeru.conf -i my_module --stop-after-init
   ```

   Install multiple modules in dependency order in one go:

   ```bash
   go run ./cmd/sumeru -- -c sumeru.conf -i base,my_module --stop-after-init
   ```

4. Start the server normally and use **Apps** or your menus to open data.

## Verify

- **`sys.module`** row for **`my_module`** with installed state
- Menus and views from **`manifest.data`** appear after sign-in (subject to access rights)

## See also

- [Update a module after XML or manifest changes](update_module.md)
- [Addons overview](../developer/addons_overview.md)
