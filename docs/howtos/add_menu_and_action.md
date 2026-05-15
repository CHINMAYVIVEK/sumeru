# How to add a menu and window action

**Goal:** open your model from the shell sidebar using a **`<menuitem>`** bound to a **`sys.action.window`** that targets your model.

## Steps

1. **Window action** — in **`views/actions.xml`** (or any **`data`** XML file), define **`sys.action.window`**:

   ```xml
   <record id="action_mymodule.document" model="sys.action.window">
       <field name="name">Documents</field>
       <field name="core_model">mymodule.document</field>
       <field name="view_mode">tree,form</field>
   </record>
   ```

2. **Views** — ensure `sys.view` rows exist for the types you listed (usually via `<view type="tree">` and `<view type="form">` in XML). Synced inline views get a default **`sys.view.name`** of `<model>.<type>` (see `core/module` view sync).

3. **Menus** — in **`views/menus.xml`**:

   ```xml
   <menuitem id="menu_mymodule_root" name="My Module" sequence="40" web_icon="apps"/>
   <menuitem id="menu_mymodule_docs" name="Documents" parent="menu_mymodule_root"
             action="mymodule.action_mymodule.document" sequence="10" web_icon="list"/>
   ```

   Use **`mymodule.`** prefix when referencing XML ids from another file in the same addon, following how **`base`** references **`base.action_core.company`**.

4. Add both XML paths to **`manifest.json`** → **`data`**, in order: actions before menus if menus reference actions.

5. Run **`-u mymodule`** (or restart with equivalent reload) so records sync.

## Verify

- Menu appears under the expected parent
- Clicking opens **tree** or **form** without 404/empty model errors

## See also

- [Views, actions, and menus](../developer/views_and_actions.md)
