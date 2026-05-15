# How to run from `sumeru_custom_addons`

**Goal:** keep **`sumeru/`** free of local addon churn while still importing **`sumeru`** as a library.

## Summary

The sibling directory **`sumeru_custom_addons/`** is a small Go module with:

- Its own **`go.mod`** and **`replace sumeru => ../sumeru`**
- A generated **`addonimports/zimports.go`** (package **`addonimports`**) listing blank imports for addons under **`sumeru_custom_addons/addons/`**
- Its own **`sumeru.conf`** (often with **`sumeru_home`** pointing at the core checkout)

## Steps (abbreviated)

1. Copy **`sumeru.conf.example`** → **`sumeru.conf`** and set database + **`addons_path`** (include both **`../sumeru/addons`** and **`addons`** if you need core apps plus local modules).

2. From **`sumeru_custom_addons/`**:

   ```bash
   make generate
   make run
   ```

3. Install modules with **`./sumeru-workspace.sh -i my_module`** or the equivalent **`go run`** documented in **`sumeru_custom_addons/README.md`**.

## See also

- **[`sumeru_custom_addons/README.md`](../../../sumeru_custom_addons/README.md)** (authoritative detail)
- [Install a module](install_module.md)
- [Tooling](../developer/tooling.md)
