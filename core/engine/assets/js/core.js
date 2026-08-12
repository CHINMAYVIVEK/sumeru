/**
 * Shell bootstrap: wires feature modules (sidebar, activity, forms, apps).
 * Each imported module owns one concern.
 */
import { initSidebar } from "./shell/sidebar.js";
import { initTopbarDropdowns } from "./shell/topbar-dropdowns.js";
import { initActivityPanel } from "./shell/activity-panel.js";
import { initNotebookTabs } from "./ui/notebook-tabs.js";
import { initAjaxFormCapture } from "./ui/ajax-form.js";
import { initAppsModuleOpener } from "./apps/apps-module-opener.js";
import { initFormSplit } from "./ui/form-split.js";
import { initMessagesComposer } from "./ui/messages-composer.js";
import { initMany2One } from "./ui/many2one.js";
import { initMany2OneSelect } from "./ui/many2one-select.js";
import { initAvatarUpload } from "./ui/avatar-upload.js";

(function bootstrap() {
  const shell = document.getElementById("sum-shell");
  if (!shell) return;

  initSidebar(shell);
  initTopbarDropdowns(shell);
  initActivityPanel(shell);
  initNotebookTabs();
  initAjaxFormCapture();
  initAppsModuleOpener();
  initFormSplit();
  initMessagesComposer();
  initMany2One();
  initMany2OneSelect();
  initAvatarUpload();
})();
