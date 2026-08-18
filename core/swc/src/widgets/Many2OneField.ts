import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../store/record.js";
import {
  fieldInputId,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";
import { AsyncFieldController } from "./field-async.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

export class Many2OneField extends SwcComponent<FieldProps> {
  private suggestions: Record<string, unknown>[] = [];
  private open = false;
  private readonly asyncCtrl = new AsyncFieldController(this);

  onWillUnmount(): void {
    this.asyncCtrl.cancel();
  }

  private async search(q: string): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const comodel = this.props.field.relation ?? this.props.field.options?.relation ?? "";
    if (!comodel) return;
    const domain = q ? [["name", "ilike", q]] : [];
    this.suggestions = await this.env.services.rpc.searchRead(comodel, domain, ["id", "name"], 20);
    this.open = true;
    this.asyncCtrl.finish(gen);
  }

  template() {
    const { field, record, readonly } = this.props;
    const display = record.get(`${field.name}_name`) ?? (record.get(field.name) ? `#${record.get(field.name)}` : "");
    const id = fieldInputId(field);

    if (readonly || field.readonly) {
      return renderFieldShell(field, fieldReadonlyValue(String(display)));
    }

    return renderFieldShell(
      field,
      html`<div class="sum-m2o-wrap">
        <input
          id=${id}
          class="sum-field-input"
          name=${field.name}
          value=${String(display)}
          autocomplete="off"
          @input=${(ev: Event) => void this.search((ev.target as HTMLInputElement).value)}
        />
        ${this.open
          ? html`<ul class="sum-m2o-suggest">
              ${this.suggestions.map(
                (row) => html`<li>
                  <button
                    type="button"
                    class="sum-m2o-option"
                    @click=${() => {
                      record.set(field.name, row.id);
                      record.set(`${field.name}_name`, row.name);
                      this.open = false;
                      this.asyncCtrl.refresh();
                    }}
                  >
                    ${String(row.name ?? row.id)}
                  </button>
                </li>`,
              )}
            </ul>`
          : ""}
      </div>`,
      { labelFor: id },
    );
  }
}
