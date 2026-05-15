# Models and fields

Addon **business models** are Go structs that implement the **`base.Model`** interface and register once at process start.

## Minimal model

1. Embed **`base.BaseModel`** (anonymous) for default ID and metadata behavior.
2. Implement **`ModelName() string`** — dotted technical name (e.g. **`core.user`**, **`sale.order`**). Follow [Naming conventions](naming_conventions.md).
3. Implement **`Fields() []base.FieldDefinition`** — column definitions the ORM uses for DDL sync and UI metadata.
4. In **`init()`**, call **`base.RegisterModel(base.RegisterModelInput{ Model: &MyModel{}, Module: "<manifest name>" })`**.

The **`Module`** string must match **`manifest.json`** **`name`** so **`sys.model`** rows and install metadata stay consistent.

Example (trimmed from **`addons/base/models/core_user.go`**):

```go
type CoreUser struct {
    base.BaseModel
    Login string `db:"login"`
    Name  string `db:"name"`
}

func (CoreUser) ModelName() string { return "core.user" }

func (CoreUser) Fields() []base.FieldDefinition {
    return []base.FieldDefinition{
        {Name: "login", Type: base.Char, Required: true, Unique: true, String: "Login", Index: true},
        {Name: "name", Type: base.Char, String: "Name"},
        // …
    }
}

func init() {
    base.RegisterModel(base.RegisterModelInput{Model: &CoreUser{}, Module: "base"})
}
```

## Field types

Declared on **`base.FieldDefinition`** (see **`sumeru/core/base/types.go`**):

- **Scalars:** `Char`, `Text`, `Integer`, `Float`, `Numeric`, `Boolean`, `Date`, `DateTime`, `Selection`, `Json`
- **Relations:** `Many2One` with **`Relation: "other.model"`**

Common field options: **`Required`**, **`Unique`**, **`Index`**, **`String`** (UI label), **`DefaultVal`**, **`Readonly`**.

Struct tags use **`db:"column_name"`** for SQL mapping. Table names are derived from the model name (e.g. **`core_user`**).

## Runtime API

Prefer **`sumeru/core/base`** helpers (struct-based inputs) over calling **`core/orm`** directly: **`Search`**, **`SearchOne`**, **`Create`**, **`Upsert`**, **`ResolveXmlId`**, etc. See **`core/base/doc.go`** for the exported surface.

## XML IDs for records

When you load `record` elements with `id="action_foo"` from your module's XML, other XML (menus, inherits) can reference `your_module.action_foo` or a short id depending on context—see `resolveXMLIDInModule` usage in menus (`action="base.action_core.company"` in `addons/base/views/menus.xml`).
