Views reference (HTML)
========================

XML views under each addon’s ``views/`` directory are loaded into ``sys.view``. The Go **render** package turns parsed views into HTML for the workspace shell.

Render package layout (``sumeru/core/engine/render/``)
------------------------------------------------------

* ``view_render.go`` — ``RenderView``: selects form / tree / kanban / pivot, builds ``PageData``, wires chatter when enabled.
* ``form_render.go`` — ``RenderForm`` and sheet/header/notebook/field/group markup.
* ``tree_render.go``, ``kanban_render.go``, ``pivot_render.go`` — list and alternate view types.
* ``workspace_chrome.go`` — workspace Edit/Save/Cancel chrome and POST form wrapper.
* ``activity_chatter_render.go`` — right-hand **Messages** tab (thread + composer).
* ``user_security_render.go`` — **Access rights** notebook tab for ``core.user``.
* ``render_types.go`` — ``PageData``, ``ViewRecordData``, ``RegisterShellHook``, ``RegisterNotebookHook``.

Hooks
-----

* **Shell**: ``RegisterShellHook`` — inject HTML into the top bar area (e.g. floating actions).
* **Notebook**: ``RegisterNotebookHook(model, pageTitle, hook)`` — replace a notebook page body for a given model and tab title.

Styles
------

Core CSS is plain ``*.css`` under ``core/engine/assets/css/``; workspace pages add ``sumeru-workspace.css``. Form/list chrome uses shared ``sum-*`` classes (see ``sumeru-views.css``). Per-page sheets (apps grid, home hub, settings hub) register extra URLs via ``ViewStylesheetURLs`` on ``PageData``.
