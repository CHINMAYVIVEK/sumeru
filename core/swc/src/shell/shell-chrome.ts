import type { HttpService } from "../services/http.js";
import { NotificationService } from "../services/notification.js";
import type { SwcBootstrap } from "../types/bootstrap.js";
import { initActivityPanel } from "./activity-panel.js";
import { initHomeDashboard } from "./home-dashboard.js";
import { initPinnedApps } from "./pinned-apps.js";
import { initSidebar } from "./sidebar.js";

export function initShellChrome(boot: SwcBootstrap, http: HttpService): void {
  const shell = document.getElementById("sum-shell");
  if (!shell) return;

  initSidebar(shell);

  if (boot.activityEnabled) {
    initActivityPanel(shell);
  }

  initPinnedApps(http, boot.pinnedApps ?? []);
  initHomeDashboard(http);

  new NotificationService().bootstrap(boot.toasts);
}
