# How to extend a view with xpath inherit

**Goal:** change another module’s stored **`sys.view.arch`** by merging an xpath fragment (see **`inherit_id`** on **`sys.view`** records in module XML).

## When to use

- You need an extra field or layout tweak on a **`base`** (or third-party) form without forking the whole XML arch.

## Steps

1. Add a **`<record>`** with **`model="sys.view"`** in your module’s XML (after the parent view exists—usually list this file **after** the module that defines the parent view in **`depends`**, or run **`-u`** twice if ordering is tricky).

2. Set fields (field names match **`sys.view`** columns loaded from XML):

   - **`inherit_id`** — XML id of the **parent** view (e.g. **`base.view_core.company_form`**). Must resolve via **`sys.model_data`**.
   - **`arch`** — fragment containing one or more **`<xpath expr="…" position="after|before|replace|inside">`** operations.

   Minimal shape:

   ```xml
   <record id="view_core.company_form_inherit_mymodule" model="sys.view">
       <field name="inherit_id" ref="base.view_core.company_form"/>
       <field name="arch" type="xml">
           <xpath expr="//field[@name='email']" position="after">
               <field name="my_extra_flag"/>
           </xpath>
       </field>
   </record>
   ```

   Adjust **`expr`** to match nodes in the **parent arch string** stored in the database (engine merges with **`viewinherit.ApplyInheritArch`**).

3. Add the file to **`manifest.json`** → **`data`**.

4. Run **`-u your_module`**.

## Caveats

- The merge runs on the **serialized arch**; xpath expressions must match that structure.
- Errors print as view-inherit warnings during sync; fix xpath and re-run **`-u`**.

## See also

- [Views, actions, and menus](../developer/views_and_actions.md)
