import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import type { SwcWorkspacePayload } from "../types/workspace.js";
import { useState } from "../core/hooks.js";
import { renderNewButton, renderReportActions, visibleFieldNames } from "./view-toolbar.js";

interface ListViewProps {
  payload: SwcWorkspacePayload;
}

export class ListView extends SwcComponent<ListViewProps> {
  private search = "";

  setup(): void {
    this.search = this.props.payload.listSearch ?? "";
    const [, bump] = useState(0);
    this.bump = () => bump((n) => n + 1);
  }

  private bump: (() => void) | null = null;

  private columns() {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
  }

  private rows() {
    return this.props.payload.records ?? [];
  }

  private applySearch(): void {
    const url = this.env.services.router.workspaceUrl({
      actionId: this.props.payload.actionId,
      menuId: this.props.payload.menuId,
      viewType: "list",
      listSearch: this.search,
    });
    this.env.services.action.navigate(url);
  }

  private openRow(row: Record<string, unknown>): void {
    const id = Number(row.id ?? 0);
    if (id <= 0) return;
    this.env.services.action.openRecord(
      this.props.payload.model,
      this.props.payload.actionId,
      this.props.payload.menuId,
      id,
      "form",
    );
  }

  template() {
    const p = this.props.payload;
    const cols = this.columns();
    const rows = this.rows();
    const fields = visibleFieldNames(cols);
    const reportActions = renderReportActions(p, fields);
    return html`
      <div class="sum-list-view">
        <div class="sum-view-toolbar">
          <div class="sum-view-toolbar-primary">
            ${renderNewButton(p)}
            <input
              type="search"
              class="sum-input sum-list-search"
              placeholder="Search…"
              value=${this.search}
              @keydown=${(ev: Event) => (ev as KeyboardEvent).key === "Enter" && this.applySearch()}
              @input=${(ev: Event) => {
                this.search = (ev.target as HTMLInputElement).value;
                this.bump?.();
              }}
            />
            <button type="button" class="sum-btn sum-btn--secondary" @click=${() => this.applySearch()}>Search</button>
          </div>
          ${reportActions ?? ""}
        </div>
        <table class="sum-list-table">
          <thead>
            <tr>${cols.map((c) => html`<th>${c.string ?? c.name}</th>`)}</tr>
          </thead>
          <tbody>
            ${rows.map(
              (row) => html`<tr class="sum-list-row" @click=${() => this.openRow(row)}>
                ${cols.map((c) => {
                  const v = row[c.name];
                  const display = row[`${c.name}_name`] ?? v;
                  return html`<td>${String(display ?? "")}</td>`;
                })}
              </tr>`,
            )}
          </tbody>
        </table>
      </div>
    `;
  }
}
