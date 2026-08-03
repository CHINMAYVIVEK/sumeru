# Models and fields

Addon **business models** are Go structs that implement the **`sdk.Model`** interface and register once at process start.

## Minimal model

1. Embed **`sdk.BaseModel`** (anonymous) for default ID and metadata behavior.
2. Implement **`ModelName() string`** — dotted technical name (e.g. **`core.user`**, **`sale.order`**). Follow [Naming conventions](naming_conventions.md).
3. Implement **`Fields() []sdk.FieldDefinition`** — column definitions the ORM uses for DDL sync and UI metadata.
4. In **`init()`**, call **`sdk.RegisterModel(sdk.RegisterModelInput{ Model: &MyModel{}, Module: "<manifest name>" })`**.

The **`Module`** string must match **`manifest.json`** **`name`** so **`sys.model`** rows and install metadata stay consistent.

Example (trimmed from **`addons/base/models/core_user.go`**):

```go
type CoreUser struct {
    sdk.BaseModel
    Login string `db:"login"`
    Name  string `db:"name"`
}

func (CoreUser) ModelName() string { return "core.user" }

func (CoreUser) Fields() []sdk.FieldDefinition {
    return []sdk.FieldDefinition{
        {Name: "login", Type: sdk.Char, Required: true, Unique: true, String: "Login", Index: true},
        {Name: "name", Type: sdk.Char, String: "Name"},
        // …
    }
}

func init() {
    sdk.RegisterModel(sdk.RegisterModelInput{Model: &CoreUser{}, Module: "base"})
}
```

## Field types

Declared on **`sdk.FieldDefinition`** (see **`sumeru/core/sdk/types.go`**):

- **Scalars:** `Char`, `Text`, `Integer`, `Float`, `Numeric`, `Boolean`, `Date`, `DateTime`, `Selection`, `Json`
- **Relations:** `Many2One` with **`Relation: "other.model"`**

Common field options: **`Required`**, **`Unique`**, **`Index`**, **`String`** (UI label), **`DefaultVal`**, **`Readonly`**.

Struct tags use **`db:"column_name"`** for SQL mapping. Table names are derived from the model name (e.g. **`core_user`**).

## Runtime API

Prefer **`sumeru/core/sdk`** helpers (struct-based inputs) over calling **`core/orm`** directly: **`Search`**, **`SearchOne`**, **`Create`**, **`Upsert`**, **`ResolveXmlId`**, etc. See **`core/sdk/doc.go`** for the exported surface.

## XML IDs for records

When you load `record` elements with `id="action_foo"` from your module's XML, other XML (menus, inherits) can reference `your_module.action_foo` or a short id depending on context—see `resolveXMLIDInModule` usage in menus (`action="base.action_core.company"` in `addons/base/views/menus.xml`).
