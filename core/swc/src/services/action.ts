import type { RouterService } from "./router.js";

export class ActionService {
  constructor(private readonly router?: RouterService) {}

  navigate(url: string): void {
    if (url.startsWith("/web?") && this.router) {
      const q = new URLSearchParams(url.slice(url.indexOf("?") + 1));
      this.router.push({
        actionId: Number(q.get("action") ?? "0"),
        menuId: q.get("menu_id") ?? "",
        viewType: q.get("view_type") ?? "",
        recordId: Number(q.get("id") ?? "0"),
        formEdit: q.get("edit") === "1",
        listSearch: q.get("q") ?? "",
      });
      return;
    }
    window.location.assign(url);
  }

  openWindowAction(actionId: number, menuId?: string, extra?: Record<string, string>): void {
    const params = new URLSearchParams({ action: String(actionId) });
    if (menuId) params.set("menu_id", menuId);
    for (const [k, v] of Object.entries(extra ?? {})) {
      if (v) params.set(k, v);
    }
    this.navigate(`/web?${params.toString()}`);
  }

  openRecord(_model: string, actionId: number, menuId: string, recordId: number, viewType = "form"): void {
    const params = new URLSearchParams({
      action: String(actionId),
      menu_id: menuId,
      view_type: viewType,
      id: String(recordId),
    });
    this.navigate(`/web?${params.toString()}`);
  }
}
