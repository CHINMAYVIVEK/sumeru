// Package module loads, installs, and updates Sumeru addons from XML manifests and data files.
//
// Install pipeline (see install.go):
//
//	discover → resolveDeps → syncSchema → loadData → syncViews → syncMenus → finalize
//
// Data sync file map:
//   - data_sync.go — orchestration, actions, generic records
//   - data_sync_views.go — sys.view inline views and inherit loading
//   - data_sync_menus.go — sys.menu from menuitem XML (deferred sync after all manifest files)
//   - data_sync_records.go — typed record upserts by model
//   - data_sync_csv.go — CSV import paths
//   - xmlid_link.go — sys.model.data XML id registry
package module
