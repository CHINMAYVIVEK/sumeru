import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { useState } from "../../runtime/hooks.js";
import { renderCollectionToolbar } from "../shared/view-toolbar.js";
import {
  renderControlPanel,
  renderRowCheckbox,
  renderSelectAllHeader,
  type ControlPanelState,
} from "./control-panel.js";
import { forEach } from "../../template/helpers.js";
import { patchKeyedChildren } from "../../runtime/patch/keyed.js";
import { VIEW_FORM, VIEW_LIST } from "../../constants/routes.js";

interface ListViewProps {
  payload: SwcWorkspacePayload;
}

export class ListView extends SwcComponent<ListViewProps> {
  private panelState: ControlPanelState = {
    search: "",
    offset: 0,
    limit: 40,
    selectedIds: new Set(),
  };

  setup(): void {
    this.panelState.search = this.props.payload.listSearch ?? "";
    const [, bump] = useState(0);
    this.bump = () => bump((n) => n + 1);
  }

  onPropsChanged(props: ListViewProps): void {
    this.panelState.search = props.payload.listSearch ?? "";
    this.panelState.offset = 0;
    this.panelState.selectedIds = new Set();
  }

  private bump: (() => void) | null = null;

  private columns() {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
  }

  private allRows() {
    let rows = [...(this.props.payload.records ?? [])];
    const order = this.panelState.order;
    if (order) {
      const desc = order.startsWith("-");
      const field = desc ? order.slice(1) : order;
      rows.sort((a, b) => {
        const av = String(a[field] ?? "");
        const bv = String(b[field] ?? "");
        return desc ? bv.localeCompare(av) : av.localeCompare(bv);
      });
    }
    return rows;
  }

  private pageRows() {
    const rows = this.allRows();
    const start = this.panelState.offset;
    return rows.slice(start, start + this.panelState.limit);
  }

  private applySearch(): void {
    this.panelState.offset = 0;
    const p = this.props.payload;
    const url = this.env.services.router.workspaceUrl({
      actionId: p.actionId,
      menuId: p.menuId,
      viewType: VIEW_LIST,
      listSearch: this.panelState.search,
    });
    this.env.services.action.navigate(url);
  }

  private applyPage(offset: number): void {
    this.panelState.offset = offset;
    this.bump?.();
  }

  private openRow(row: Record<string, unknown>): void {
    const id = Number(row.id ?? 0);
    if (id <= 0) return;
    this.env.services.action.openRecord({
      actionId: this.props.payload.actionId,
      menuId: this.props.payload.menuId,
      recordId: id,
      viewType: VIEW_FORM,
    });
  }

  private toggleRow(id: number, checked: boolean): void {
    if (checked) this.panelState.selectedIds.add(id);
    else this.panelState.selectedIds.delete(id);
    this.bump?.();
  }

  private toggleAll(checked: boolean, ids: number[]): void {
    this.panelState.selectedIds = checked ? new Set(ids) : new Set();
    this.bump?.();
  }

  private async bulkDelete(): Promise<void> {
    const ids = [...this.panelState.selectedIds];
    if (ids.length === 0) return;
    const ok = await this.env.services.dialog.confirm(
      "Delete records",
      `Delete ${ids.length} selected record(s)?`,
    );
    if (!ok) return;
    await this.env.services.rpc.unlink(this.props.payload.model, ids);
    this.panelState.selectedIds = new Set();
    this.env.services.notification.show({
      kind: "success",
      title: "Deleted",
      body: `${ids.length} record(s) removed.`,
    });
    this.applySearch();
  }

  private renderRow(row: Record<string, unknown>) {
    const id = Number(row.id ?? 0);
    const cols = this.columns();
    return html`<tr class="sum-list-row sum-list-row--click" @click=${() => this.openRow(row)}>
      ${renderRowCheckbox(id, this.panelState.selectedIds.has(id), (rid, checked) =>
        this.toggleRow(rid, checked),
      )}
      ${cols.map((c) => {
        const display = row[`${c.name}_name`] ?? row[c.name];
        return html`<td class="sum-list-td">${String(display ?? "")}</td>`;
      })}
    </tr>`;
  }

  patch(): void {
    const tbody = this.el?.querySelector("tbody");
    if (tbody) {
      const rows = this.pageRows();
      patchKeyedChildren(
        tbody,
        rows.map((row) => ({
          key: String(row.id ?? 0),
          render: () => this.renderRow(row).render(),
        })),
      );
      return;
    }
    super.patch();
  }

  template() {
    const p = this.props.payload;
    const cols = this.columns();
    const rows = this.pageRows();
    const allRows = this.allRows();
    const ids = allRows.map((r) => Number(r.id ?? 0)).filter((id) => id > 0);
    const allSelected = ids.length > 0 && ids.every((id) => this.panelState.selectedIds.has(id));

    return html`
      <div class="sum-list-view">
        ${renderCollectionToolbar({
          payload: p,
          viewType: VIEW_LIST,
          search: this.panelState.search,
          onSearch: () => this.applySearch(),
          onInput: (next) => {
            this.panelState.search = next;
          },
          extraPrimary:
            this.panelState.selectedIds.size > 0
              ? html`<button type="button" class="sum-btn sum-btn--danger" @click=${() => void this.bulkDelete()}>
                  Delete (${this.panelState.selectedIds.size})
                </button>`
              : "",
        })}
        ${renderControlPanel({
          payload: { ...p, records: allRows },
          state: this.panelState,
          onPage: (o) => this.applyPage(o),
        })}
        <div class="sum-list-table-wrap">
          <table class="sum-list-table">
            <thead>
              <tr>
                ${renderSelectAllHeader(allSelected, (checked) => this.toggleAll(checked, ids))}
                ${cols.map((c) => html`<th class="sum-list-th">${c.string ?? c.name}</th>`)}
              </tr>
            </thead>
            <tbody>
              ${forEach(rows, (row) => Number(row.id ?? 0), (row) => this.renderRow(row))}
            </tbody>
          </table>
        </div>
      </div>
    `;
  }
}
