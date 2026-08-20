import {
  Q_ACTION,
  Q_EDIT,
  Q_MENU_ID,
  Q_RECORD_ID,
  Q_SEARCH,
  Q_VIEW_TYPE,
  VIEW_FORM,
  WEB_ROUTE,
  EDIT_ENABLED,
} from "../constants/routes.js";
import { RouterService } from "./router.js";

export type OpenRecordOpts = {
  actionId: number;
  menuId: string;
  recordId: number;
  viewType?: string;
};

export class ActionService {
  constructor(private readonly router?: RouterService) {}

  navigate(url: string): void {
    if (url.startsWith(`${WEB_ROUTE}?`) && this.router) {
      const q = new URLSearchParams(url.slice(url.indexOf("?") + 1));
      this.router.push({
        actionId: Number(q.get(Q_ACTION) ?? "0"),
        menuId: q.get(Q_MENU_ID) ?? "",
        viewType: q.get(Q_VIEW_TYPE) ?? "",
        recordId: Number(q.get(Q_RECORD_ID) ?? "0"),
        formEdit: q.get(Q_EDIT) === EDIT_ENABLED,
        listSearch: q.get(Q_SEARCH) ?? "",
      });
      return;
    }
    window.location.assign(url);
  }

  openWindowAction(actionId: number, menuId?: string, extra?: Record<string, string>): void {
    const params = new URLSearchParams({ [Q_ACTION]: String(actionId) });
    if (menuId) params.set(Q_MENU_ID, menuId);
    for (const [k, v] of Object.entries(extra ?? {})) {
      if (v) params.set(k, v);
    }
    this.navigate(`${WEB_ROUTE}?${params.toString()}`);
  }

  openRecord({ actionId, menuId, recordId, viewType = VIEW_FORM }: OpenRecordOpts): void {
    const route = {
      actionId,
      menuId,
      recordId,
      viewType,
      formEdit: false,
      listSearch: "",
    };
    if (this.router) {
      this.router.push(route);
      return;
    }
    this.navigate(RouterService.buildUrl(route));
  }
}
