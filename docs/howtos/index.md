# How-tos

Short **task recipes**: goal → steps → verification. For background concepts, start at **[`docs/developer/`](../developer/index.md)**.

| Recipe | When to use |
| ------ | ----------- |
| [Scaffold a new addon](scaffold_new_addon.md) | Greenfield module with manifest, models, views, security |
| [Install a module](install_module.md) | First-time load of an addon into a database |
| [Update a module after XML or manifest changes](update_module.md) | Changed views, menus, data XML, or `manifest.json` |
| [Add a model and register it](add_model.md) | New Go struct + ORM table |
| [Add a menu and window action](add_menu_and_action.md) | Open a model from the shell |
| [Extend a view with xpath inherit](view_inherit_xpath.md) | Patch another module’s arch |
| [Run from `sumeru_custom_addons`](custom_addons_workspace.md) | Keep core repo clean; local INI and imports |
| [Troubleshoot common dev issues](troubleshooting.md) | Blank imports, wrong DB, menu not showing |
