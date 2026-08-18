import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../store/record.js";
import {
  fieldInputId,
  fieldPlaceholder,
  fieldReadonlyInput,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

function inputTypeForField(field: SwcArchField): string {
  if (field.widget === "email") return "email";
  if (field.type === "integer" || field.type === "float" || field.type === "numeric") return "number";
  if (field.type === "date") return "date";
  if (field.type === "datetime") return "datetime-local";
  return "text";
}

function stepForField(field: SwcArchField): string | undefined {
  if (field.type === "integer") return "1";
  if (field.type === "float" || field.type === "numeric") return "any";
  return undefined;
}

function parseNumericValue(field: SwcArchField, raw: string): unknown {
  if (raw === "") return null;
  if (field.type === "integer") return Number.parseInt(raw, 10);
  if (field.type === "float" || field.type === "numeric") return Number.parseFloat(raw);
  return raw;
}

export class DefaultField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const val = String(record.get(field.name) ?? "");
    const placeholder = fieldPlaceholder(field);
    const inputType = inputTypeForField(field);
    const step = stepForField(field);
    const id = fieldInputId(field);

    if (readonly || field.readonly) {
      return renderFieldShell(
        field,
        field.type === "integer" || field.type === "float" || field.type === "numeric"
          ? fieldReadonlyInput(field, val, "text")
          : fieldReadonlyValue(val, placeholder),
      );
    }

    return renderFieldShell(
      field,
      html`<input
        id=${id}
        type=${inputType}
        class="sum-field-input"
        name=${field.name}
        placeholder=${placeholder}
        value=${val}
        ${step ? html`step=${step}` : ""}
        @input=${(ev: Event) =>
          record.set(field.name, parseNumericValue(field, (ev.target as HTMLInputElement).value))}
      />`,
      { labelFor: id },
    );
  }
}
