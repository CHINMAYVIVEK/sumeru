# How to update a module after XML or manifest changes

**Goal:** reload **`manifest.data`** (views, menus, security XML, demo records) from disk without reinstalling from scratch.

## When to use **`-u`**

Run an update after you change any of:

- Files listed under **`manifest.json`** → **`data`**
- **`manifest.json`** itself (**`depends`**, **`version`**, etc.)
- Go model **`Fields()`** (schema sync runs on startup; **`-u`** refreshes metadata paths—see below)

## Steps

1. Restart is not always required for pure XML if your process reloads on next request, but the supported workflow is via CLI:

   ```bash
   go run ./cmd/sumeru -- -c sumeru.conf -u my_module --stop-after-init
   ```

2. Or update everything installed:

   ```bash
   go run ./cmd/sumeru -- -c sumeru.conf -u all --stop-after-init
   ```

3. Start the server again (or omit **`--stop-after-init`** if you want update + HTTP in one process—see **`sumeru/README.md`**).

## Schema (Go fields)

Table alterations run when the server initializes the ORM; changing **`Fields()`** typically needs a normal server start (or any run that initializes DB). Use **`-u`** when you need **module metadata** and **XML** reloaded reliably.

## See also

- [Install a module](install_module.md)
- [Tooling](../developer/tooling.md)
