import { SwcApp } from "./core/app.js";
import { SwcEnv } from "./core/env.js";
import { readBootstrap } from "./types/bootstrap.js";
import { RpcService } from "./services/rpc.js";
import { HttpService } from "./services/http.js";
import { NotificationService } from "./services/notification.js";
import { ActionService } from "./services/action.js";
import { RouterService } from "./services/router.js";
import { BusService } from "./services/bus.js";
import { ShellLayout } from "./shell/ShellLayout.js";
import { initShellChrome } from "./shell/shell-chrome.js";
import { registerDefaultWidgets } from "./widgets/registry.js";
import { registry, type RegistryEntry } from "./core/registry.js";
import { AddonLoader } from "./addon/loader.js";
import { ListView } from "./views/ListView.js";
import { FormView } from "./views/FormView.js";
import { KanbanView } from "./views/KanbanView.js";
import { PivotView } from "./views/PivotView.js";
import { GraphView } from "./views/GraphView.js";

function registerCore(): void {
  registerDefaultWidgets();
  const views = registry.category("views");
  views.add("list", ListView as unknown as RegistryEntry);
  views.add("form", FormView as unknown as RegistryEntry);
  views.add("kanban", KanbanView as unknown as RegistryEntry);
  views.add("pivot", PivotView as unknown as RegistryEntry);
  views.add("graph", GraphView as unknown as RegistryEntry);
  const main = registry.category("main_components");
  main.add("shell", ShellLayout as unknown as RegistryEntry);
}

function buildEnv(boot: ReturnType<typeof readBootstrap>): SwcEnv {
  const services = {
    rpc: new RpcService(boot.rpcUrl, boot.csrfToken),
    http: new HttpService(boot.csrfToken),
    notification: new NotificationService(),
    action: new ActionService(),
    router: new RouterService(),
    bus: new BusService(),
  };
  return new SwcEnv(boot, services);
}

function bootstrap(): void {
  registerCore();
  AddonLoader.registerFromGlobal();

  let boot;
  try {
    boot = readBootstrap();
  } catch {
    return;
  }

  const env = buildEnv(boot);
  initShellChrome(boot, env.services.http);

  const mountEl = document.getElementById("swc-workspace");
  if (mountEl) {
    SwcApp.start(mountEl, env, ShellLayout);
  }
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", bootstrap);
} else {
  bootstrap();
}

export { SwcApp, registry, SwcEnv };
