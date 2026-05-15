# Views, actions, and menus

Sumeru loads declarative UI from **XML** files listed in **`manifest.json`** → **`data`**. Root element is **`<sumeru>`** with **`<data>`** wrapping records, views, and menus.

## Window actions (`sys.action.window`)

Define an action that opens a model in the workspace:

```xml
<record id="action_my.model" model="sys.action.window">
    <field name="name">My Records</field>
    <field name="core_model">my_module.my_model</field>
    <field name="view_mode">tree,form</field>
    <field name="help">Optional help text.</field>
</record>
```

- **`core_model`** — target model technical name (required on window actions).
- **`view_mode`** — comma-separated order of view types (e.g. **`tree,kanban,form`**).

## Inline views (`<view>`)

Primary form/tree/kanban definitions use **`<view id="…" model="…" type="form|tree|kanban|pivot">`**. The engine stores a generated **`arch`** string in **`sys.view`** and maps **`module.xml_id`** to that row.

Structure (simplified):

- **Form:** **`<sheet>`**, **`<group>`**, **`<notebook>`** / **`<page>`**, **`<field name="…"/>`**
- **Tree:** **`<field>`** columns; optional **`open`** attribute on **`<view type="tree">`** to control row→form navigation
- **Chatter (optional):** **`<chatter>`** with **`<field widget="mail_thread"/>`** for the activity panel

Study **`addons/base/views/core_company_form_views.xml`** and **`core_company_tree_views.xml`** as canonical examples.

## View inheritance (`sys.view` records)

Extensions are **`<record>`** rows with **`model="sys.view"`**, **`inherit_id`** pointing to the parent view’s XML id, and an **`arch`** fragment containing **`<xpath expr="…" position="after|before|replace|inside">`** nodes. Inherits are applied **after** all primary views in the same sync pass (`inheritQueue` in **`core/module/data_sync.go`**).

Use this when you must alter another module’s arch without copying the whole view.

## Menus (`<menuitem>`)

Parsed from the same XML files (see **`parser.ParseMenus`**):

```xml
<menuitem id="menu_my_root" name="My App" sequence="10" web_icon="apps"/>
<menuitem id="menu_my_items" name="Items" parent="menu_my_root" sequence="10"
          action="my_module.action_my.model" web_icon="list"/>
```

- **`parent`** — XML id of parent menu (module-qualified or same-file id per resolver rules).
- **`action`** — XML id of a **`sys.action.window`** record.
- **`access_groups`** — optional; limits visibility to users in those groups (XML id list / resolution matches shell menu loading).

Pinned **Home** / **Settings** roots in **`base`** are defined in **`addons/base/views/menus.xml`**.

## Further reading

- Render pipeline and hook points: [Web layer](web_layer.md) and **`docs/refrence/views.rst`**.
- [Web Icons](icons.md) — Reference for `web_icon` and adding new symbols.
