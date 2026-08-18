import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import type { SwcWorkspacePayload } from "../types/workspace.js";
import { ListView } from "./ListView.js";
import { FormView } from "./FormView.js";
import { KanbanView } from "./KanbanView.js";
import { PivotView } from "./PivotView.js";
import { GraphView } from "./GraphView.js";
import { useState, useEffect } from "../core/hooks.js";
import { SwcError } from "../core/error.js";

export class WorkspaceRouter extends SwcComponent {
  private payload: SwcWorkspacePayload | null = null;
  private loading = true;
  private error = "";

  setup(): void {
    const [, bump] = useState(0);
    this.bump = () => bump((n) => n + 1);

    const load = async (): Promise<void> => {
      this.loading = true;
      this.error = "";
      this.bump?.();
      try {
        this.payload = await this.fetchWorkspace();
      } catch (err) {
        this.error = err instanceof SwcError ? err.message : String(err);
      } finally {
        this.loading = false;
        this.bump?.();
      }
    };

    void load();
    useEffect(() => {
      const onNav = (): void => void load();
      window.addEventListener("popstate", onNav);
      return () => window.removeEventListener("popstate", onNav);
    });
  }

  private bump: (() => void) | null = null;

  private async fetchWorkspace(): Promise<SwcWorkspacePayload> {
    const route = this.env.services.router.parse();
    const params = new URLSearchParams();
    if (route.actionId) params.set("action", String(route.actionId));
    if (route.menuId) params.set("menu_id", route.menuId);
    if (route.viewType) params.set("view_type", route.viewType);
    if (route.recordId) params.set("id", String(route.recordId));
    if (route.formEdit) params.set("edit", "1");
    if (route.listSearch) params.set("q", route.listSearch);
    const base = this.env.bootstrap.swcApiBase || "/web/swc";
    return this.env.services.http.getJSON(`${base}/workspace?${params.toString()}`);
  }

  private mountView(view: { setup?: () => void; render(): HTMLElement }): HTMLElement {
    view.setup?.();
    return view.render();
  }

  private renderView(): HTMLElement {
    if (!this.payload) return document.createElement("div");
    const type = this.payload.viewType || this.payload.arch.type;
    const p = this.payload;
    switch (type) {
      case "form":
        return this.mountView(new FormView({ payload: p }, this.env));
      case "kanban":
        return this.mountView(new KanbanView({ payload: p }, this.env));
      case "pivot":
        return this.mountView(new PivotView({ payload: p }, this.env));
      case "graph":
        return this.mountView(new GraphView({ payload: p }, this.env));
      default:
        return this.mountView(new ListView({ payload: p }, this.env));
    }
  }

  template() {
    if (this.loading) {
      return html`<div class="sum-workspace-loading">Loading workspace…</div>`;
    }
    if (this.error) {
      return html`<div class="sum-flash sum-flash--error">${this.error}</div>`;
    }
    return html`<div class="sum-workspace-view">${this.renderView()}</div>`;
  }
}
