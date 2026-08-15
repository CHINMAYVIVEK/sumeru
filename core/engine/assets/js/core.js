/**
 * Shell bootstrap: wires feature modules (sidebar, activity, forms).
 * Each imported module owns one concern.
 */
import { initSidebar } from "./shell/sidebar.js";
import { initTopbarDropdowns } from "./shell/topbar-dropdowns.js";
import { initActivityPanel } from "./shell/activity-panel.js";
import { initAppLauncher } from "./shell/app-launcher.js";
import { applyTopNavFilter, initRecentTracking, initPinnedApps } from "./shell/pinned-apps.js";
import { initNotebookTabs } from "./ui/notebook-tabs.js";
import { initMessagesComposer } from "./ui/messages-composer.js";
import { initMany2One } from "./ui/many2one.js";
import { initMany2OneSelect } from "./ui/many2one-select.js";
import { initMultiSelect } from "./ui/multi-select.js";
import { initAvatarUpload } from "./ui/avatar-upload.js";
import { initKanbanBoard } from "./ui/kanban-board.js";
import { initStatusbar } from "./ui/statusbar.js";
import { initPriorityField } from "./ui/priority-field.js";
import { initReportExchange } from "./ui/report-exchange.js";
import { initBulkImportForm } from "./ui/bulk-import.js";
import { initFlashDismiss } from "./ui/flash-dismiss.js";
import { initFormValidation } from "./ui/form-validation.js";
import { initWorkspaceToastsFromDOM } from "./lib/toast.js";

export {
  getPinnedApps,
  getRecentApps,
  togglePinnedApp,
  pushRecentApp,
  applyTopNavFilter,
} from "./shell/pinned-apps.js";

(function bootstrap() {
  const shell = document.getElementById("sum-shell");
  if (!shell) return;

  initSidebar(shell);
  initTopbarDropdowns(shell);
  initActivityPanel(shell);
  initNotebookTabs();
  initMessagesComposer();
  initMany2One();
  initMany2OneSelect();
  initMultiSelect();
  initAvatarUpload();
  initKanbanBoard();
  initStatusbar();
  initPriorityField();
  initReportExchange();
  initBulkImportForm();
  initFlashDismiss();
  initFormValidation();
  initWorkspaceToastsFromDOM();
  initRecentTracking();
  initPinnedApps();
  applyTopNavFilter();
  initAppLauncher();
})();
