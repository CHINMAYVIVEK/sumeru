import { html, type TemplateResult } from "../../template/html.js";
import type { SwcArchField, SwcWorkspacePayload } from "../../types/workspace.js";

function linkButton(href: string, label: string, className = "sum-btn sum-btn--secondary"): HTMLElement {
  const a = document.createElement("a");
  a.className = className;
  a.href = href;
  a.textContent = label;
  return a;
}

export function visibleFieldNames(fields: SwcArchField[]): string {
  return fields
    .map((f) => f.name)
    .filter(Boolean)
    .join(",");
}

export function newRecordUrl(payload: SwcWorkspacePayload): string {
  const params = new URLSearchParams();
  if (payload.actionId > 0) params.set("action", String(payload.actionId));
  if (payload.menuId) params.set("menu_id", payload.menuId);
  params.set("view_type", "form");
  return `/web?${params.toString()}`;
}

export function editRecordUrl(payload: SwcWorkspacePayload): string {
  if (payload.formBaseQuery) {
    return `/web?${payload.formBaseQuery}&edit=1`;
  }
  const params = new URLSearchParams();
  if (payload.actionId > 0) params.set("action", String(payload.actionId));
  if (payload.menuId) params.set("menu_id", payload.menuId);
  params.set("view_type", "form");
  if (payload.recordId > 0) params.set("id", String(payload.recordId));
  params.set("edit", "1");
  return `/web?${params.toString()}`;
}

export function exportQuery(
  payload: SwcWorkspacePayload,
  fields: string,
  recordId = 0,
): URLSearchParams {
  const params = new URLSearchParams();
  params.set("model", payload.model);
  if (payload.actionId > 0) params.set("action", String(payload.actionId));
  if (fields) params.set("fields", fields);
  if (recordId > 0) params.set("id", String(recordId));
  return params;
}

export function renderNewButton(payload: SwcWorkspacePayload): HTMLElement {
  return linkButton(newRecordUrl(payload), "New", "sum-btn sum-list-btn-new");
}

export function toolbarButton(
  label: string,
  className: string,
  onClick: () => void,
  disabled = false,
): HTMLElement {
  const btn = document.createElement("button");
  btn.type = "button";
  btn.className = className;
  btn.textContent = label;
  btn.disabled = disabled;
  btn.addEventListener("click", onClick);
  return btn;
}

/** Maps arch button class (e.g. sum_highlight) to pre-SWC header button modifiers. */
export function resolveHeaderButtonClass(archClass?: string): string {
  const base = "sum-header-btn";
  if (archClass?.includes("sum_highlight")) {
    return `${base} sum-header-btn--primary`;
  }
  return `${base} sum-header-btn--secondary`;
}

export function headerButton(
  label: string,
  archClass: string | undefined,
  onClick: () => void,
  disabled = false,
): HTMLElement {
  const className = disabled
    ? `${resolveHeaderButtonClass(archClass)} sum-header-btn--disabled`
    : resolveHeaderButtonClass(archClass);
  return toolbarButton(label, className, onClick, disabled);
}

export function renderReportActions(
  payload: SwcWorkspacePayload,
  fields: string,
  recordId = 0,
): HTMLElement | null {
  const report = payload.arch.report;
  if (!report?.download && !report?.upload) return null;

  const exportParams = exportQuery(payload, fields, recordId);
  const items: Array<HTMLElement | TemplateResult> = [];

  if (report.download) {
    items.push(linkButton(`/web/export/csv?${exportParams.toString()}`, "Export CSV"));
    items.push(linkButton(`/web/export/pdf?${exportParams.toString()}`, "Export PDF"));
  }
  if (report.upload && fields) {
    const templateParams = new URLSearchParams(exportParams);
    items.push(linkButton(`/web/bulk/template?${templateParams.toString()}`, "Download template"));
    items.push(
      html`<form class="sum-list-upload-form" method="post" enctype="multipart/form-data" action="/web/bulk/upload">
        <input type="hidden" name="csrf_token" value=${payload.csrfToken} />
        <input type="hidden" name="model" value=${payload.model} />
        ${payload.actionId > 0 ? html`<input type="hidden" name="action" value=${String(payload.actionId)} />` : ""}
        <input type="hidden" name="fields" value=${fields} />
        <label class="sum-btn sum-btn--secondary sum-list-upload-label">
          Import CSV
          <input type="file" name="file" accept=".csv,text/csv" class="sum-list-upload-input" @change=${(ev: Event) => (ev.target as HTMLInputElement).form?.requestSubmit()} />
        </label>
      </form>`,
    );
  }
  if (items.length === 0) return null;

  const wrap = document.createElement("div");
  wrap.className = "sum-view-toolbar-actions";
  for (const item of items) {
    wrap.appendChild(item instanceof HTMLElement ? item : item.render());
  }
  return wrap;
}
