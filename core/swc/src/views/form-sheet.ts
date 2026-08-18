import { html, type TemplateResult } from "../core/template.js";
import type { SwcEnv } from "../core/env.js";
import type {
  SwcArchField,
  SwcArchGroup,
  SwcArchNotebook,
  SwcArchSheet,
  SwcArchDiv,
  SwcArchSeparator,
  SwcArchLabel,
} from "../types/workspace.js";
import type { SwcRecord } from "../store/record.js";
import { renderField as defaultRenderField } from "../widgets/registry.js";

export type RenderFieldFn = (
  field: SwcArchField,
  record: SwcRecord,
  readonly: boolean,
) => HTMLElement;

function renderFields(
  rf: RenderFieldFn,
  fields: SwcArchField[],
  record: SwcRecord,
  readonly: boolean,
): Array<TemplateResult | HTMLElement> {
  return visibleFields(fields).map((f) => rf(f, record, readonly));
}

function collectDivFields(div: SwcArchDiv): SwcArchField[] {
  const out = [...(div.fields ?? []), ...(div.h1Fields ?? [])];
  for (const nested of div.divs ?? []) {
    out.push(...collectDivFields(nested));
  }
  return out;
}

export function collectFormFields(sheet?: SwcArchSheet, headerFields: SwcArchField[] = []): SwcArchField[] {
  const out = [...headerFields];
  if (!sheet) return out.filter((f) => !f.invisible);

  out.push(...(sheet.fields ?? []));
  for (const div of sheet.divs ?? []) {
    out.push(...collectDivFields(div));
  }
  for (const g of sheet.groups ?? []) {
    out.push(...collectGroupFields(g));
  }
  for (const nb of sheet.notebook ?? []) {
    for (const pg of nb.pages ?? []) {
      out.push(...(pg.fields ?? []));
      for (const g of pg.groups ?? []) {
        out.push(...collectGroupFields(g));
      }
    }
  }
  return out.filter((f) => !f.invisible);
}

function collectGroupFields(group: SwcArchGroup): SwcArchField[] {
  const out = [...(group.fields ?? [])];
  for (const nested of group.groups ?? []) {
    out.push(...collectGroupFields(nested));
  }
  return out;
}

function visibleFields(fields: SwcArchField[]): SwcArchField[] {
  return fields.filter((f) => !f.invisible);
}

function renderSeparators(separators: SwcArchSeparator[] = []): TemplateResult {
  if (separators.length === 0) return html``;
  return html`${separators.map((sep) =>
    sep.string
      ? html`<div class="sum-separator--title">${sep.string}</div>`
      : html`<hr class="sum-separator--rule" />`,
  )}`;
}

function renderLabels(labels: SwcArchLabel[] = []): TemplateResult {
  if (labels.length === 0) return html``;
  return html`${labels.map((lab) => html`<div class="sum-label--notebook">${lab.string ?? ""}</div>`)}`;
}

function initialsFromName(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

function renderHeroField(
  field: SwcArchField,
  record: SwcRecord,
  readonly: boolean,
): TemplateResult {
  const val = String(record.get(field.name) ?? "");
  const placeholder = field.placeholder ?? field.string ?? field.name;
  if (readonly || field.readonly) {
    return html`<h1><div class="sum-form-hero-input sum-form-hero-input--bold">${val}</div></h1>`;
  }
  return html`<h1>
    <input
      class="sum-form-hero-input sum-form-hero-input--bold"
      name=${field.name}
      placeholder=${placeholder}
      value=${val}
      @input=${(ev: Event) => record.set(field.name, (ev.target as HTMLInputElement).value)}
    />
  </h1>`;
}

function renderContactItem(
  field: SwcArchField,
  record: SwcRecord,
  readonly: boolean,
): TemplateResult {
  const val = String(record.get(field.name) ?? "");
  const label = field.string ?? field.name;
  const inputType = field.widget === "email" ? "email" : "text";
  if (readonly || field.readonly) {
    return html`<div class="sum-form-contact-item">
      <label class="sum-field-label">${label}</label>
      <div class="sum-form-inline-input">${val}</div>
    </div>`;
  }
  return html`<div class="sum-form-contact-item">
    <label class="sum-field-label">${label}</label>
    <input
      type=${inputType}
      class="sum-form-inline-input"
      name=${field.name}
      placeholder=${label}
      value=${val}
      @input=${(ev: Event) => record.set(field.name, (ev.target as HTMLInputElement).value)}
    />
  </div>`;
}

function renderAvatar(record: SwcRecord, readonly: boolean): TemplateResult {
  const image = String(record.get("image") ?? "");
  const name = String(record.get("name") ?? "");
  const hasImage = image.length > 0;
  const initials = initialsFromName(name);

  return html`<div class="sum-form-avatar sum-form-avatar--compact" data-sum-avatar>
    <div class="sum-form-avatar-box sum-form-avatar-box--circle">
      ${hasImage
        ? html`<img
            .sum-form-avatar-img
            .sum-form-avatar-img--visible
            class=${image.includes("data:") ? "sum-form-avatar-img--cropped" : ""}
            src=${image}
            alt=""
          />`
        : html`<span class="sum-form-avatar-initials">${initials}</span>`}
    </div>
    ${readonly
      ? ""
      : html`<div class="sum-form-avatar-actions">
          <input
            type="hidden"
            name="image"
            data-sum-avatar-value
            value=${image}
            @input=${(ev: Event) => record.set("image", (ev.target as HTMLInputElement).value)}
          />
          <label class="sum-form-avatar-upload">
            Upload
            <input type="file" accept="image/*" />
          </label>
        </div>`}
  </div>`;
}

function renderTitleBody(
  rf: RenderFieldFn,
  div: SwcArchDiv,
  record: SwcRecord,
  readonly: boolean,
): TemplateResult {
  const h1Fields = visibleFields(div.h1Fields ?? []);
  const contactDiv = (div.divs ?? []).find((d) => (d.class ?? "").includes("sum-title-contact-row"));
  const contactFields = visibleFields(contactDiv?.fields ?? []);

  return html`<div class="sum-form-title-body sum-form-title-body--main">
    ${h1Fields.length > 0 ? renderHeroField(h1Fields[0], record, readonly) : ""}
    ${contactFields.length > 0
      ? html`<div class="sum-title-contact-row">
          ${contactFields.map((f) => renderContactItem(f, record, readonly))}
        </div>`
      : ""}
    ${h1Fields.length === 0 && contactFields.length === 0
      ? renderFields(rf, div.fields ?? [], record, readonly)
      : ""}
  </div>`;
}

function renderTitleDiv(
  rf: RenderFieldFn,
  div: SwcArchDiv,
  record: SwcRecord,
  readonly: boolean,
  hasImageField: boolean,
): TemplateResult {
  const cls = div.class ?? "";
  if (cls.includes("sum_button_box")) {
    return html`<div class="sum-form-button-box ${cls}"></div>`;
  }

  const isTitle = cls.includes("sum_title");
  if (!isTitle) {
    return html`<div class=${cls}>${renderFields(rf, div.fields ?? [], record, readonly)}</div>`;
  }

  const h1Fields = visibleFields(div.h1Fields ?? []);
  const legacySingle = h1Fields.length === 0 && visibleFields(div.fields ?? []).length === 1;
  const titleField = h1Fields[0] ?? (legacySingle ? visibleFields(div.fields ?? [])[0] : undefined);

  if (hasImageField) {
    return html`<div class="sum-form-split-layout sum-form-split-layout--compact" data-sum-form-split>
      <aside class="sum-form-split-left sum-form-split-left--avatar">${renderAvatar(record, readonly)}</aside>
      <div class="sum-form-split-main">${renderTitleBody(rf, div, record, readonly)}</div>
    </div>`;
  }

  if (titleField) {
    return html`<div class="sum-form-title-row sum-form-title-row--sheet">
      ${renderTitleBody(rf, div, record, readonly)}
    </div>`;
  }

  return html`<div class="sum-form-title-row sum-form-title-row--sheet">
    ${renderTitleBody(rf, div, record, readonly)}
  </div>`;
}

function renderGroup(
  rf: RenderFieldFn,
  group: SwcArchGroup,
  record: SwcRecord,
  readonly: boolean,
  plain = false,
): TemplateResult {
  const nested = group.groups ?? [];
  const fields = group.fields ?? [];
  const hasNested = nested.length > 0;
  const groupModifier = plain || !group.string ? "sum-form-group--plain" : "sum-form-group--full";

  if (hasNested && fields.length === 0) {
    return html`<div class="sum-form-edit-grid sum-field-region--sheet">
      ${nested.map((g) => renderGroup(rf, g, record, readonly))}
    </div>`;
  }

  return html`<div .sum-form-group class=${groupModifier}>
    ${group.string ? html`<div class="sum-form-group-title">${group.string}</div>` : ""}
    <div class="sum-form-group-grid">
      ${renderFields(rf, fields, record, readonly)}
      ${renderSeparators(group.separators)}
      ${renderLabels(group.labels)}
      ${nested.map((g) => renderGroup(rf, g, record, readonly, true))}
    </div>
  </div>`;
}

function renderNotebook(
  rf: RenderFieldFn,
  notebook: SwcArchNotebook,
  record: SwcRecord,
  readonly: boolean,
  notebookIndex: number,
  activePage: number,
  onTab: (notebookIndex: number, pageIndex: number) => void,
): TemplateResult {
  const pages = notebook.pages ?? [];
  if (pages.length === 0) return html``;

  const idx = Math.min(Math.max(activePage, 0), pages.length - 1);
  const page = pages[idx];

  return html`<div class="sum-notebook sum-notebook--sheet">
    <div class="sum-notebook-tabs" role="tablist">
      ${pages.map((pg, i) => {
        const tabClass = i === idx ? "sum-notebook-tab sum-notebook-tab--active" : "sum-notebook-tab";
        return html`<button type="button" class=${tabClass} role="tab" aria-selected=${i === idx ? "true" : "false"} @click=${() => onTab(notebookIndex, i)}>${pg.title}</button>`;
      })}
    </div>
    <div class="sum-notebook-page sum-notebook-page--sheet" role="tabpanel">
      <div class="sum-form-page-grid">
        ${renderFields(rf, page.fields ?? [], record, readonly)}
        ${(page.groups ?? []).map((g) => renderGroup(rf, g, record, readonly))}
        ${renderSeparators(page.separators)}
        ${renderLabels(page.labels)}
      </div>
    </div>
  </div>`;
}

export interface RenderFormSheetOptions {
  env: SwcEnv;
  sheet?: SwcArchSheet;
  record: SwcRecord;
  readonly: boolean;
  hasImageField?: boolean;
  activeNotebookPages: Record<number, number>;
  onNotebookTab: (notebookIndex: number, pageIndex: number) => void;
  renderField?: RenderFieldFn;
}

export function renderFormSheet(opts: RenderFormSheetOptions): TemplateResult {
  const {
    env,
    sheet,
    record,
    readonly,
    hasImageField = false,
    activeNotebookPages,
    onNotebookTab,
    renderField: renderFieldOpt,
  } = opts;
  const rf: RenderFieldFn =
    renderFieldOpt ?? ((f, r, ro) => defaultRenderField(env, f, r, ro));
  if (!sheet) {
    return html`<div class="sum-form-sheet"></div>`;
  }

  const parts: Array<TemplateResult | HTMLElement> = [];

  for (const div of sheet.divs ?? []) {
    parts.push(renderTitleDiv(rf, div, record, readonly, hasImageField));
  }

  const topFields = visibleFields(sheet.fields ?? []);
  const groups = sheet.groups ?? [];
  if (topFields.length > 0 || groups.length > 0) {
    parts.push(
      html`<div class="sum-form-edit-grid sum-field-region--sheet">
        ${renderFields(rf, topFields, record, readonly)}
        ${groups.map((g) => renderGroup(rf, g, record, readonly))}
      </div>`,
    );
  }

  (sheet.notebook ?? []).forEach((nb, notebookIndex) => {
    const activePage = activeNotebookPages[notebookIndex] ?? 0;
    parts.push(renderNotebook(rf, nb, record, readonly, notebookIndex, activePage, onNotebookTab));
  });

  const sheetSeparators = sheet.separators ?? [];
  const sheetLabels = sheet.labels ?? [];
  if (sheetSeparators.length > 0 || sheetLabels.length > 0) {
    parts.push(
      html`<div class="sum-form-sheet-meta">${renderSeparators(sheetSeparators)}${renderLabels(sheetLabels)}</div>`,
    );
  }

  return html`<div class="sum-form-sheet">${parts}</div>`;
}
