import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcArchField, SwcWorkspacePayload } from "../../types/workspace.js";
import { renderNewButton, renderReportActions, visibleFieldNames } from "../shared/view-toolbar.js";
import { renderKanbanCardInner } from "./kanban-card.js";

interface KanbanViewProps {
  payload: SwcWorkspacePayload;
}

export class KanbanView extends SwcComponent<KanbanViewProps> {
  private cardFields(): SwcArchField[] {
    return this.props.payload.arch.fields.filter((f) => !f.invisible);
  }

  private openCard(row: Record<string, unknown>): void {
    const id = Number(row.id ?? 0);
    if (id <= 0) return;
    const p = this.props.payload;
    this.env.services.action.openRecord(p.model, p.actionId, p.menuId, id, "form");
  }

  private async moveCard(recordId: number, columnValue: number): Promise<void> {
    const groupField = this.props.payload.arch.kanban?.groupField;
    if (!groupField) return;
    await this.env.services.rpc.write(this.props.payload.model, [recordId], {
      [groupField]: columnValue || false,
    });
    this.env.services.bus.emit("record.updated", {
      model: this.props.payload.model,
      id: recordId,
    });
  }

  private toolbar() {
    const p = this.props.payload;
    const fields = visibleFieldNames(this.cardFields());
    const reportActions = renderReportActions(p, fields);
    return html`
      <div class="sum-view-toolbar sum-kanban-report-bar">
        <div class="sum-view-toolbar-primary">${renderNewButton(p)}</div>
        ${reportActions ?? ""}
      </div>
    `;
  }

  private renderCard(row: Record<string, unknown>, fields: SwcArchField[], draggable = false) {
    return html`<div
      class="sum-kanban-card"
      draggable=${draggable ? "true" : undefined}
      @click=${() => this.openCard(row)}
      @dragstart=${draggable
        ? (ev: Event) => (ev as DragEvent).dataTransfer?.setData("text/plain", String(row.id))
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
                    html`<div
                      class="sum-kanban-card"
                      draggable=${kanban.draggable ? "true" : undefined}
                      @click=${() => this.openCard(row)}
                      @dragstart=${(ev: Event) => (ev as DragEvent).dataTransfer?.setData("text/plain", String(row.id))}
                      @dragover=${(ev: Event) => ev.preventDefault()}
                      @drop=${(ev: Event) => {
                        ev.preventDefault();
                        const de = ev as DragEvent;
                        const id = Number(de.dataTransfer?.getData("text/plain"));
                        if (id) void this.moveCard(id, col.value);
                      }}
                    >
                      ${renderKanbanCardInner(row, fields)}
                    </div>`,
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
