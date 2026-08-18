import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../store/record.js";
import {
  fieldInputId,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

export class TextareaField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const val = String(record.get(field.name) ?? "");
    const placeholder = field.placeholder ?? field.string ?? field.name;
    const id = fieldInputId(field);

    if (readonly || field.readonly) {
      return renderFieldShell(field, fieldReadonlyValue(val));
    }

    return renderFieldShell(
      field,
      html`<textarea
        id=${id}
        class="sum-field-textarea"
        name=${field.name}
        placeholder=${placeholder}
        rows="4"
        @input=${(ev: Event) => record.set(field.name, (ev.target as HTMLTextAreaElement).value)}
      >
${val}</textarea
      >`,
      { labelFor: id },
    );
  }
}
