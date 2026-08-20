import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { VIEW_FORM } from "../../constants/routes.js";

interface CalendarViewProps {
  payload: SwcWorkspacePayload;
}

export class CalendarView extends SwcComponent<CalendarViewProps> {
  private dateField = "date_deadline";

  setup(): void {
    const fields = this.props.payload.arch.fields;
    const dateField = fields.find((f) => f.type === "date" || f.type === "datetime");
    if (dateField) this.dateField = dateField.name;
  }

  private groupByDate(): Map<string, Record<string, unknown>[]> {
    const map = new Map<string, Record<string, unknown>[]>();
    for (const row of this.props.payload.records ?? []) {
      const raw = String(row[this.dateField] ?? "").slice(0, 10);
      const key = raw || "Unscheduled";
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(row);
    }
    return map;
  }

  private openRecord(row: Record<string, unknown>): void {
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

  template() {
    const buckets = [...this.groupByDate().entries()].sort(([a], [b]) => a.localeCompare(b));
    return html`
      <div class="sum-calendar-view">
        <h2 class="sum-calendar-title">${this.props.payload.arch.title ?? "Calendar"}</h2>
        <div class="sum-calendar-columns">
          ${buckets.map(
            ([day, rows]) => html`<section class="sum-calendar-day">
              <h3 class="sum-calendar-day-title">${day}</h3>
              <ul class="sum-calendar-events">
                ${rows.map(
                  (row: Record<string, unknown>) => html`<li class="sum-calendar-event" @click=${() => this.openRecord(row)}>
                    ${String(row.name ?? row.display_name ?? `#${row.id}`)}
                  </li>`,
                )}
              </ul>
            </section>`,
          )}
        </div>
      </div>
    `;
  }
}
