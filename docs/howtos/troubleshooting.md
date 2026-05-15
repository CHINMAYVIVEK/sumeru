# Troubleshooting common development issues

## “My new Go model is ignored” / table never appears

- Confirm **`base.RegisterModel(..., Module: "<manifest name>")`** uses the **exact** **`manifest.json`** **`name`**.
- Ensure the addon root package is blank-imported in **`cmd/sumeru/zimports.go`** (or **`addonimports/zimports.go`** in **`sumeru_custom_addons`**): run **`make generate`**.
- Rebuild/restart the server so **`init()`** runs.

## “Module not found” or wrong addon wins

- **`addons_path`** order: **later** directories override duplicate **`manifest.name`**.
- Paths are resolved relative to the **INI file directory** unless absolute.

## “I changed XML but UI is stale”

- Run **`-u my_module`** (or **`-u all`**) so **`manifest.data`** reloads. See [Update a module](update_module.md).

## “Menu never shows up”

- Check **`access_groups`** on **`menuitem`** — user may lack required groups.
- Confirm **`action="module.xml_id"`** resolves to a real **`sys.action.window`** row.
- Ensure the module is **installed**, not only discovered.

## “View inherit failed” / warning on startup

- **`inherit_id`** must resolve to an existing **`sys.view`** row (parent loaded first; fix **`depends`** / XML order).
- XPath must match the **stored parent arch** string. Test smaller fragments; read the warning from **`applySysUIViewInherit`**.

## Database confusion

- Remember **`-d`** / **`--database`** overrides **`db_name`** for a single process—easy to inspect the “wrong” DB.

## Still stuck

- Run **`go test ./...`** in **`sumeru/`** on a clean tree to ensure local changes did not break parsing or render.
- Search **`core/module`** and **`core/engine/parser`** for the feature you are extending; errors often print with **`platformmsg`** format strings.
