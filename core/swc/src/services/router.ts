export interface WorkspaceRoute {
  actionId: number;
  menuId: string;
  viewType: string;
  recordId: number;
  formEdit: boolean;
  listSearch: string;
}

export class RouterService {
  parse(location: Location = window.location): WorkspaceRoute {
    const q = new URLSearchParams(location.search);
    return {
      actionId: Number(q.get("action") ?? "0"),
      menuId: q.get("menu_id") ?? "",
      viewType: q.get("view_type") ?? "",
      recordId: Number(q.get("id") ?? "0"),
      formEdit: q.get("edit") === "1",
      listSearch: q.get("q") ?? "",
    };
  }

  workspaceUrl(route: Partial<WorkspaceRoute>): string {
    const current = this.parse();
    const merged = { ...current, ...route };
    const params = new URLSearchParams();
    if (merged.actionId) params.set("action", String(merged.actionId));
    if (merged.menuId) params.set("menu_id", merged.menuId);
    if (merged.viewType) params.set("view_type", merged.viewType);
    if (merged.recordId) params.set("id", String(merged.recordId));
    if (merged.formEdit) params.set("edit", "1");
    if (merged.listSearch) params.set("q", merged.listSearch);
    return `/web?${params.toString()}`;
  }

  push(route: Partial<WorkspaceRoute>): void {
    const url = this.workspaceUrl(route);
    window.history.pushState({}, "", url);
    window.dispatchEvent(new PopStateEvent("popstate"));
  }
}
