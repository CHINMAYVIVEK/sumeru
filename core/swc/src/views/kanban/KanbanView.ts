import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcArchField, SwcWorkspacePayload } from "../../types/workspace.js";
import { renderCollectionToolbar } from "../shared/view-toolbar.js";
import { renderKanbanCardInner } from "./kanban-card.js";
import { RECORD_UPDATED, VIEW_FORM, VIEW_KANBAN } from "../../constants/routes.js";

interface KanbanViewProps {
  payload: SwcWorkspacePayload;
}

export class KanbanView extends SwcComponent<KanbanViewProps> {
  private search = "";

  setup(): void {
    this.search = this.props.payload.listSearch ?? "";
  }

  onPropsChanged(props: KanbanViewProps): void {
    this.search = props.payload.listSearch ?? "";
  }

  private cardFields(): SwcArchField[] {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
  }

  private applySearch(): void {
    const p = this.props.payload;
    this.env.services.action.navigate(
      this.env.services.router.workspaceUrl({
        actionId: p.actionId,
        menuId: p.menuId,
        viewType: VIEW_KANBAN,
        listSearch: this.search,
      }),
    );
  }

  private openCard(row: Record<string, unknown>): void {
    const id = Number(row.id ?? 0);
    if (id <= 0) return;
    const p = this.props.payload;
    this.env.services.action.openRecord({
      actionId: p.actionId,
      menuId: p.menuId,
      recordId: id,
      viewType: VIEW_FORM,
    });
  }

  private async moveCard(recordId: number, columnValue: number): Promise<void> {
    const groupField = this.props.payload.arch.kanban?.groupField;
    if (!groupField) return;
    await this.env.services.rpc.write(this.props.payload.model, [recordId], {
      [groupField]: columnValue || false,
    });
    this.env.services.bus.emit(RECORD_UPDATED, {
      model: this.props.payload.model,
      id: recordId,
    });
  }

  private toolbar() {
    return renderCollectionToolbar({
      payload: this.props.payload,
      viewType: VIEW_KANBAN,
      search: this.search,
      onSearch: () => this.applySearch(),
      onInput: (next) => {
        this.search = next;
      },
    });
  }

  private renderCard(
    row: Record<string, unknown>,
    fields: SwcArchField[],
    opts: { draggable?: boolean; dropValue?: number } = {},
  ) {
    const draggable = Boolean(opts.draggable);
    const dropValue = opts.dropValue;
    return html`<div
      class="sum-kanban-card"
      draggable=${draggable ? "true" : undefined}
      @click=${() => this.openCard(row)}
      @dragstart=${draggable
        ? (ev: Event) => (ev as DragEvent).dataTransfer?.setData("text/plain", String(row.id))
        : undefined}
      @dragover=${dropValue !== undefined ? (ev: Event) => ev.preventDefault() : undefined}
      @drop=${dropValue !== undefined
        ? (ev: Event) => {
            ev.preventDefault();
            const id = Number((ev as DragEvent).dataTransfer?.getData("text/plain"));
            if (id) void this.moveCard(id, dropValue);
          }
        : undefined}
    >
      ${renderKanbanCardInner(row, fields)}
    </div>`;
  }

  template() {
    const p = this.props.payload;
    const kanban = p.arch.kanban;
    const fields = this.cardFields();
    if (!kanban?.columns?.length) {
      const rows = p.records ?? [];
      return html`
        <div class="sum-kanban-view">
          ${this.toolbar()}
          <div class="sum-kanban-columns">
            ${rows.length === 0
              ? html`<div class="sum-kanban-empty">No records</div>`
              : rows.map((row) => this.renderCard(row, fields))}
          </div>
        </div>
      `;
    }
    return html`
      <div class="sum-kanban-view">
        ${this.toolbar()}
        <div class="sum-kanban-board sum-kanban-board--grouped">
          <div class="sum-kanban-stage-columns">
            ${kanban.columns.map(
              (col) => html`<div class="sum-kanban-stage-column" data-column=${String(col.value)}>
                <div class="sum-kanban-stage-header">
                  <span>${col.label}</span>
                  <span class="sum-kanban-stage-count">${String(col.records.length)}</span>
                </div>
                <div class="sum-kanban-cards">
                  ${col.records.map((row) =>
                    this.renderCard(row, fields, { draggable: kanban.draggable, dropValue: col.value }),
                  )}
                </div>
              </div>`,
            )}
          </div>
        </div>
      </div>
    `;
  }
}
