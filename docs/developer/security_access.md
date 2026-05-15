# Security and access

Sumeru separates **who** can use the UI (**groups**), **what CRUD** they have per model (**access control lists**), and optional **row-level** filters (**record rules**).

## Groups (`core.group`)

Groups bundle users for menus and ACLs. Each group has a **name**, optional **`category_id`** (many2one to **`sys.module.category`** for UI grouping), **sequence**, and **implied** links stored in **`core.group.implied`** (loaded from XML **`implied_ids`** eval `[(4, ref('xml.id')), …]`). Kernel groups **`base.group_system`** and **`base.group_user`** are created before module data install so addons can imply **`base.group_user`**. Admin UI for groups lives under **Settings** when **`security_admin.xml`** from **`base`** is loaded.

## Access rights (`sys.access`)

Each row ties **(optional) group + model** to **`perm_read`**, **`perm_write`**, **`perm_create`**, **`perm_unlink`** booleans.

### CSV bulk load

Place **`sys.access.csv`** at the addon root **or** under **`security/sys.access.csv`**. Columns are read by **`core/module`** during **`SyncToDB`** (see **`syncCSVModelAccess`**). Use this for repeatable ACL matrices in bulk CSV form.

### XML records

You can also declare **`sys.access`** (and related) rows as **`<record model="sys.access">`** blocks in security XML—see **`addons/base/security/security_admin.xml`** for patterns.

## Record rules (`sys.rule`)

Optional domains (JSON) restrict rows per operation. Form views expose **`domain_force`** for administrators (see **`view_sys.rule_form`** in **`security_admin.xml`**).

## Menu visibility

**`<menuitem access_groups="…">`** ties shell entries to groups resolved at menu load time (`LoadShellMenus` in **`core/engine/render/menus.go`**).

## Practical workflow

1. Define **groups** your module needs.
2. Ship **`security/sys.access.csv`** (or XML) so each new model has sane defaults.
3. Add **record rules** only when users must not see each other’s rows.
4. Restrict sensitive **Settings** menus with **`access_groups`** matching **`settings.group_system`** (see **`base`** security XML).
