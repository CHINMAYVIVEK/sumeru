/**
 * Shell bootstrap: wires feature modules (sidebar, activity, forms, apps).
 * Each imported module owns one concern.
 */
import { initSidebar } from "./shell/sidebar.js";
import { initActivityPanel } from "./shell/activity-panel.js";
import { initNotebookTabs } from "./ui/notebook-tabs.js";
import { initAjaxFormCapture } from "./ui/ajax-form.js";
import { initAppsModuleOpener } from "./apps/apps-module-opener.js";

(function bootstrap() {
  const shell = document.getElementById("sum-shell");
  if (!shell) return;

  initSidebar(shell);
  initActivityPanel(shell);
  initNotebookTabs();
  initAjaxFormCapture();
  initAppsModuleOpener();
})();
