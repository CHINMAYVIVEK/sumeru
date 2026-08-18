export class ActionService {
  navigate(url: string): void {
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
