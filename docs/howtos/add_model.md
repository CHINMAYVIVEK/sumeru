# How to add a model and register it

**Goal:** new PostgreSQL table + **`sys.model`** row + optional UI later.

## Steps

1. Under **`sumeru/addons/<your_module>/models/`**, add **`my_model.go`**:

   - Struct embeds **`base.BaseModel`**
   - **`ModelName()`** returns a dotted name (e.g. **`mymodule.document`**)
   - **`Fields()`** lists columns
   - **`init()`** calls **`sdk.RegisterModel(..., Module: "<manifest name>")`**

2. Ensure **`addons/<your_module>/init.go`** blank-imports **`sumeru/addons/<your_module>/models`**.

3. Run **`make generate`** so **`cmd/sumeru`** links your addon package.

4. Start the server (or **`-i`/`-u`**) so ORM sync creates/updates the table and **`sys.model`** picks up the declaring module.

## Verify

- Table exists in PostgreSQL (name derived from model, e.g. **`mymodule_document`**)
- **`sys.model`** contains your model with correct **`module`** column

## Next

- [Add a menu and window action](add_menu_and_action.md)
- [Security](../developer/security_access.md)
