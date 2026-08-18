import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../store/record.js";
import { renderFieldShell } from "./field-shell.js";
import { AsyncFieldController } from "./field-async.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

function inverseFieldName(parentModel: string): string {
  const part = parentModel.split(".").pop() ?? "parent";
  return `${part}_id`;
}

export class One2ManyField extends SwcComponent<FieldProps> {
  private lines: Record<string, unknown>[] = [];
  private loaded = false;
  private readonly asyncCtrl = new AsyncFieldController(this);

  setup(): void {
    void this.loadLines();
  }

  onWillUnmount(): void {
    this.asyncCtrl.cancel();
  }

  private async loadLines(): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const { field, record } = this.props;
    const comodel = field.relation ?? field.options?.relation ?? "";
    if (!comodel || record.id <= 0) {
      this.loaded = true;
      this.asyncCtrl.finish(gen);
      return;
    }
    const inv = field.options?.inverse ?? inverseFieldName(record.model);
    this.lines = await this.env.services.rpc.searchRead(
      comodel,
      [[inv, "=", record.id]],
      ["name", "quantity", "unit_price", "note"],
      200,
    );
    this.loaded = true;
    this.asyncCtrl.finish(gen);
  }

  template() {
    const { field } = this.props;
    const label = field.string ?? field.name;

    return renderFieldShell(
      field,
      html`<div class="sum-o2m-table-wrap">
        <div class="sum-o2m-title">${label}</div>
        <table class="sum-o2m-table">
          <thead>
            <tr>
              <th>Description</th>
              <th>Qty</th>
              <th>Unit price</th>
              <th>Note</th>
            </tr>
          </thead>
          <tbody>
            ${this.lines.length === 0
              ? html`<tr><td colspan="4">${this.loaded ? "No lines" : "Loading…"}</td></tr>`
              : this.lines.map(
                  (line) => html`<tr>
                    <td>${String(line.name ?? "")}</td>
                    <td>${String(line.quantity ?? "")}</td>
                    <td>${String(line.unit_price ?? "")}</td>
                    <td>${String(line.note ?? "")}</td>
                  </tr>`,
                )}
          </tbody>
        </table>
      </div>`,
    );
  }
}
