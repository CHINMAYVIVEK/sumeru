import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../store/record.js";
import { renderFieldShell } from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

export class PriorityField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const label = field.string ?? field.name;
    const stars = 3;
    const current = Number(record.get(field.name) ?? 0);

    return renderFieldShell(
      field,
      html`<div class="sum-priority-stars" role="group" aria-label=${label}>
        ${Array.from({ length: stars }, (_, i) => i + 1).map((n) => {
          const starClass = n <= current ? "sum-priority-star sum-priority-star--on" : "sum-priority-star";
          return html`<button
            type="button"
            class=${starClass}
            disabled=${readonly ? "disabled" : undefined}
            aria-label=${`Priority ${n}`}
            @click=${() => !readonly && record.set(field.name, n)}
          >
            ★
          </button>`;
        })}
      </div>`,
      { showLabel: true },
    );
  }
}
