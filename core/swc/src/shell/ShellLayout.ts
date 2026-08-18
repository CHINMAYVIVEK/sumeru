import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import { useState, useEffect } from "../core/hooks.js";
import { AppLauncher } from "./AppLauncher.js";
import { WorkspaceRouter } from "../views/WorkspaceRouter.js";

export class ShellLayout extends SwcComponent {
  private launcherOpen = false;
  private workspaceRouter!: WorkspaceRouter;
  private appLauncher!: AppLauncher;

  setup(): void {
    const [, setTick] = useState(0);
    const bump = (): void => setTick((n) => n + 1);

    this.workspaceRouter = new WorkspaceRouter({}, this.env);
    this.workspaceRouter.setup?.();

    const boot = this.env.bootstrap;
    this.appLauncher = new AppLauncher(
      {
        apps: boot.apps,
        isOpen: () => this.launcherOpen,
        requestClose: () => {
          this.launcherOpen = false;
          bump();
        },
      },
      this.env,
    );
    this.appLauncher.setup?.();

    useEffect(() => {
      const toggleLauncher = (): void => {
        this.launcherOpen = !this.launcherOpen;
        bump();
      };
      const onKey = (ev: KeyboardEvent): void => {
        if ((ev.ctrlKey || ev.metaKey) && ev.key.toLowerCase() === "k") {
          ev.preventDefault();
          toggleLauncher();
        }
      };
      document.addEventListener("keydown", onKey);
      document.addEventListener("swc:launcher-toggle", toggleLauncher);
      const searchBtn = document.getElementById("sum-topbar-search-open");
      searchBtn?.addEventListener("click", toggleLauncher);
      return () => {
        document.removeEventListener("keydown", onKey);
        document.removeEventListener("swc:launcher-toggle", toggleLauncher);
        searchBtn?.removeEventListener("click", toggleLauncher);
      };
    });

    if (this.env.bootstrap.busEnabled) {
      this.env.services.bus.connect();
    }
  }

  private workspaceView(): HTMLElement {
    if (this.workspaceRouter.el?.isConnected) {
      this.workspaceRouter.patch();
      return this.workspaceRouter.el;
    }
    return this.workspaceRouter.render();
  }

  template() {
    return html`
      <div id="swc-root-inner">
        <main class="sum-workspace-inner">
          ${this.workspaceView()}
        </main>
        ${this.appLauncher.render()}
      </div>
    `;
  }
}
