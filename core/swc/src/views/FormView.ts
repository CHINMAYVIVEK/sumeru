import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import type { SwcArchButton, SwcWorkspacePayload } from "../types/workspace.js";
import { RecordStore, SwcRecord } from "../store/record.js";
import { useState } from "../core/hooks.js";
import { SwcError } from "../core/error.js";
import {
  headerButton,
  renderNewButton,
  renderReportActions,
  visibleFieldNames,
} from "./view-toolbar.js";
import { collectFormFields, renderFormSheet } from "./form-sheet.js";
import { initFormInteractions } from "./form-interactions.js";
import { FieldHost } from "../widgets/field-host.js";

interface FormViewProps {
  payload: SwcWorkspacePayload;
}

export class FormView extends SwcComponent<FormViewProps> {
  private recordStore!: RecordStore;
  private record!: SwcRecord;
  private snapshot: Record<string, unknown> = {};
  private editing = false;
  private saving = false;
  private acting = false;
  private error = "";
  private activeNotebookPages: Record<number, number> = {};
  private teardownInteractions: (() => void) | null = null;
  private fieldHost!: FieldHost;

  setup(): void {
    const [, bump] = useState(0);
    this.bump = () => bump((n) => n + 1);
    this.recordStore = new RecordStore(this.env.services.rpc);
    this.fieldHost = new FieldHost(this.env);
    const p = this.props.payload;
    this.editing = p.formEdit || p.recordId <= 0;
    this.snapshot = { ...(p.record ?? {}) };
    this.record = this.recordStore.fromPayload(p.model, p.recordId, this.snapshot);
  }

  private bump: (() => void) | null = null;

  onMount(): void {
    if (this.el) {
      this.teardownInteractions?.();
      this.teardownInteractions = initFormInteractions(this.el);
    }
  }

  onWillUnmount(): void {
    this.teardownInteractions?.();
    this.teardownInteractions = null;
    this.fieldHost.clear();
  }

  private renderFieldCached = (
    field: import("../types/workspace.js").SwcArchField,
    record: SwcRecord,
    readonly: boolean,
  ): HTMLElement => this.fieldHost.render(field, record, readonly);

  private isReadonly(): boolean {
    return !this.editing;
  }

  private toolbarBusy(): boolean {
    return this.saving || this.acting;
  }

  private fields() {
    const arch = this.props.payload.arch;
    return collectFormFields(arch.sheet, arch.header?.fields ?? []);
  }

  private headerButtons(): SwcArchButton[] {
    return this.props.payload.arch.header?.buttons ?? [];
  }

  private startEdit(): void {
    this.editing = true;
    this.error = "";
    this.bump?.();
  }

  private cancelEdit(): void {
    const p = this.props.payload;
    if (p.recordId <= 0) {
      const url = this.env.services.router.workspaceUrl({
        actionId: p.actionId,
        menuId: p.menuId,
        viewType: "list",
        recordId: 0,
        formEdit: false,
      });
      this.env.services.action.navigate(url);
      return;
    }
    this.record = this.recordStore.fromPayload(p.model, p.recordId, { ...this.snapshot });
    this.editing = false;
    this.error = "";
    this.bump?.();
  }

  private async reloadRecord(): Promise<void> {
    const p = this.props.payload;
    if (p.recordId <= 0) return;
    const fieldNames = this.fields().map((f) => f.name);
    const rows = await this.env.services.rpc.read(p.model, [p.recordId], fieldNames);
    if (!rows[0]) return;
    this.snapshot = { ...rows[0] };
    this.record = this.recordStore.fromPayload(p.model, p.recordId, this.snapshot);
    this.bump?.();
  }

  private async save(): Promise<void> {
    this.saving = true;
    this.error = "";
    this.bump?.();
    try {
      const required = this.fields().filter((f) => f.required).map((f) => f.name);
      this.recordStore.validate(this.record, required);
      const id = await this.recordStore.save(this.record);
      this.env.services.notification.show({
        kind: "success",
        title: "Saved",
        body: "Record saved successfully.",
      });
      const p = this.props.payload;
      if (p.recordId <= 0 && id > 0) {
        this.env.services.action.openRecord(p.model, p.actionId, p.menuId, id, "form");
        return;
      }
      this.snapshot = { ...this.record.data };
      this.editing = false;
      this.bump?.();
    } catch (err) {
      this.error = err instanceof SwcError ? err.message : String(err);
    } finally {
      this.saving = false;
      this.bump?.();
    }
  }

  private async runObjectButton(btn: SwcArchButton): Promise<void> {
    const p = this.props.payload;
    if (btn.type !== "object" || p.recordId <= 0) return;
    this.acting = true;
    this.error = "";
    this.bump?.();
    try {
      const result = (await this.env.services.rpc.callMethod(p.model, btn.name, p.recordId)) as {
        redirect?: string;
      };
      if (result?.redirect) {
        this.env.services.action.navigate(result.redirect);
        return;
      }
      this.env.services.notification.show({
        kind: "success",
        title: btn.string || btn.name,
        body: "Action completed.",
      });
      await this.reloadRecord();
    } catch (err) {
      this.error = err instanceof SwcError ? err.message : String(err);
    } finally {
      this.acting = false;
      this.bump?.();
    }
  }

  private renderToolbarPrimary(): Array<HTMLElement> {
    const p = this.props.payload;
    const busy = this.toolbarBusy();
    const items: HTMLElement[] = [];

    if (p.recordId > 0 && this.isReadonly()) {
      items.push(renderNewButton(p));
      items.push(headerButton("Edit", undefined, () => this.startEdit(), busy));
    } else {
      items.push(
        headerButton("Save", "sum_highlight", () => void this.save(), busy),
      );
      items.push(headerButton("Cancel", undefined, () => this.cancelEdit(), busy || this.saving));
    }

    for (const btn of this.headerButtons()) {
      if (btn.type !== "object") continue;
      items.push(
        headerButton(btn.string || btn.name, btn.class, () => void this.runObjectButton(btn), busy),
      );
    }

    return items;
  }

  template() {
    const p = this.props.payload;
    const readonly = this.isReadonly();
    const headerFields = p.arch.header?.fields ?? [];
    const exportFields = visibleFieldNames(this.fields());
    const reportActions =
      p.recordId > 0 ? renderReportActions(p, exportFields, p.recordId) : null;
    const toolbarItems = this.renderToolbarPrimary();
    const busy = this.toolbarBusy();

    const sheet = renderFormSheet({
      env: this.env,
      sheet: p.arch.sheet,
      record: this.record,
      readonly,
      hasImageField: p.arch.formMeta?.hasImageField ?? false,
      activeNotebookPages: this.activeNotebookPages,
      onNotebookTab: (notebookIndex, pageIndex) => {
        this.activeNotebookPages = { ...this.activeNotebookPages, [notebookIndex]: pageIndex };
        this.bump?.();
      },
      renderField: this.renderFieldCached,
    });

    const footerButtons = p.arch.footer?.buttons ?? [];

    return html`
      <div class="sum-form-view sum-form-view--workspace-chrome${readonly ? " sum-form-view--readonly" : ""}">
        <div class="sum-ws-record-toolbar sum-view-toolbar sum-form-toolbar">
          <div class="sum-statusbar-buttons sum-view-toolbar-primary">${toolbarItems}</div>
          ${headerFields.length > 0
            ? html`<div class="sum-statusbar-status sum-ws-toolbar-right">
                ${headerFields.map((f) => this.renderFieldCached(f, this.record, readonly))}
              </div>`
            : ""}
          ${reportActions ?? ""}
        </div>
        ${this.error ? html`<div class="sum-flash sum-flash--error">${this.error}</div>` : ""}
        <div class="sum-form-sheet-bg">
          ${sheet}
          ${footerButtons.length > 0
            ? html`<div class="sum-form-footer">
                ${footerButtons.map((btn) =>
                  headerButton(
                    btn.string || btn.name,
                    btn.class,
                    () => void this.runObjectButton(btn),
                    busy,
                  ),
                )}
              </div>`
            : ""}
        </div>
      </div>
    `;
  }
}
