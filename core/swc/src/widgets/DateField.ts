import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../store/record.js";
import {
  fieldInputId,
  fieldPlaceholder,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

function isDateTime(field: SwcArchField): boolean {
  return field.type === "datetime" || field.widget === "datetime";
}

function toNativeValue(field: SwcArchField, raw: unknown): string {
  const text = String(raw ?? "").trim();
  if (!text) return "";
  if (isDateTime(field)) {
    const d = new Date(text);
    if (Number.isNaN(d.getTime())) return text.slice(0, 16);
    const pad = (n: number) => String(n).padStart(2, "0");
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }
  return text.slice(0, 10);
}

function formatDisplay(field: SwcArchField, raw: unknown): string {
  const native = toNativeValue(field, raw);
  if (!native) return "";
  if (isDateTime(field)) {
    const d = new Date(native);
    if (Number.isNaN(d.getTime())) return native;
    return d.toLocaleString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }
  const d = new Date(`${native}T00:00:00`);
  if (Number.isNaN(d.getTime())) return native;
  return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

function todayNative(field: SwcArchField): string {
  const d = new Date();
  const pad = (n: number) => String(n).padStart(2, "0");
  if (isDateTime(field)) {
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

function closeDetails(ev: Event): void {
  const details = (ev.currentTarget as HTMLElement | null)?.closest("details.sum-date-field");
  if (details instanceof HTMLDetailsElement) details.open = false;
}

export class DateField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const raw = record.get(field.name);
    const native = toNativeValue(field, raw);
    const display = formatDisplay(field, raw);
    const placeholder = fieldPlaceholder(field);
    const id = fieldInputId(field);
    const inputType = isDateTime(field) ? "datetime-local" : "date";

    if (readonly || field.readonly) {
      return renderFieldShell(field, fieldReadonlyValue(display, placeholder));
    }

    return renderFieldShell(
      field,
      html`<details class="sum-date-field">
        <summary class="sum-date-field-trigger">
          <span class=${display ? "sum-date-field-value" : "sum-date-field-value sum-date-field-value--placeholder"}>
            ${display || placeholder}
          </span>
          <span class="sum-date-field-icon" aria-hidden="true">📅</span>
        </summary>
        <input type="hidden" id=${id} name=${field.name} value=${native} />
        <div class="sum-date-popover" role="dialog" aria-label=${placeholder}>
          <div class="sum-date-popover-header">${field.string ?? field.name}</div>
          <input
            type=${inputType}
            class="sum-date-popover-input"
            value=${native}
            @input=${(ev: Event) => {
              record.set(field.name, (ev.target as HTMLInputElement).value || null);
              this.patch();
            }}
            @change=${(ev: Event) => {
              record.set(field.name, (ev.target as HTMLInputElement).value || null);
              this.patch();
            }}
          />
          <div class="sum-date-popover-actions">
            <button
              type="button"
              class="sum-date-popover-btn"
              @click=${(ev: Event) => {
                record.set(field.name, todayNative(field));
                this.patch();
                closeDetails(ev);
              }}
            >
              Today
            </button>
            <button
              type="button"
              class="sum-date-popover-btn"
              @click=${(ev: Event) => {
                record.set(field.name, null);
                this.patch();
                closeDetails(ev);
              }}
            >
              Clear
            </button>
            <button type="button" class="sum-date-popover-btn sum-date-popover-btn--primary" @click=${closeDetails}>
              Done
            </button>
          </div>
        </div>
      </details>`,
      { labelFor: id },
    );
  }
}

export { DateField as DateTimeField };
