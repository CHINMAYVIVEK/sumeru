# Addons overview

An **addon** (module) is a directory that contains **`manifest.json`** and optional **`models/`**, **`views/`**, **`security/`**, and **`static/`** assets.

## Discovery and naming

- Each **root** in **`addons_path`** (comma-separated in `sumeru.conf`) is scanned for **first-level subdirectories** that contain **`manifest.json`**.
- **`manifest.name`** is the technical module name. It must match **`^[a-z][a-z0-9_]*$`** (same rule as **`sumeru-bp`**).
- If the same **`name`** appears in two roots, the **later** path in **`addons_path`** **wins** (override pattern for forks or `sumeru_custom_addons`).

## `manifest.json`

| Field | Notes |
| ----- | ----- |
| `name` | Technical name; equals folder name in normal setups |
| `display_name` | Shown in Apps / UI when set |
| `version` | Semantic string for humans and **`sys.module`** |
| `depends` | List of module names; topological order for install/update |
| `author`, `description` | Metadata |
| `data` | Ordered list of **XML paths relative to the addon root** loaded on install/update |
| `application` | If **`true`** or omitted, module can appear as an application in Apps; set **`false`** for technical glue modules |

Only **installed** modules have their **`data`** XML applied on startup sync. Uninstalled modules are still **discovered** (visible in Apps) but menus/views from them are not loaded.

## Go side-effect imports

Go structs register with the ORM in **`init()`** via **`sdk.RegisterModel`**. The **`cmd/sumeru`** binary must **blank-import** the addon root package so those **`init`** hooks run.

- In-repo default: **`go generate ./cmd/sumeru`** runs **`sumeru-import-gen`**, which rewrites **`cmd/sumeru/zimports.go`** from **`addons_path`** in your INI.
- After adding a **new** addon under a configured root, run **`make generate`** (or **`go generate ./cmd/sumeru`**) then **`-i your_module`** once.

See [Tooling](tooling.md) and [How-to: install a module](../howtos/install_module.md).

## Typical layout

```
addons/my_module/
  manifest.json
  init.go                 # usually: import _ "sumeru/addons/my_module/models"
  models/
    my_model.go           # RegisterModel(..., Module: "my_module")
  views/
    actions.xml
    menus.xml
    my_model_form.xml
  security/
    security.xml          # optional XML records
    sys.access.csv        # optional CSV ACLs
  static/
    css/theme-overrides.css   # optional; served as /static/addon-css/<module>.css
```

The reference implementation is **`sumeru/addons/base/`**.
