import { html, type TemplateResult } from "../core/template.js";
import type { SwcArchField } from "../types/workspace.js";

export function fieldInputId(field: SwcArchField): string {
  return `f-${field.name}`;
}

function isFullWidthField(field: SwcArchField): boolean {
  if (field.type === "text" || field.widget === "text") return true;
  if (field.type === "one2many" || field.widget === "one2many") return true;
  if (field.widget === "image") return true;
  return false;
}

export function fieldWidgetClass(field: SwcArchField, extra: string[] = []): string {
  const parts = ["sum-field-widget"];
  if (isFullWidthField(field)) {
    parts.push("sum-field-widget--full");
  }
  if (field.type === "many2one" || field.widget === "many2one") {
    parts.push("sum-field-widget--many2one");
  }
  for (const mod of extra) {
    if (mod) parts.push(mod);
  }
  return parts.join(" ");
}

export function fieldLabel(field: SwcArchField, forId?: string, row = false): TemplateResult {
  const label = field.string ?? field.name;
  const cls = row ? "sum-field-label sum-field-label--row" : "sum-field-label";
  if (forId) {
    return html`<label class=${cls} for=${forId}>${label}</label>`;
  }
  return html`<label class=${cls}>${label}</label>`;
}

export function fieldControl(body: TemplateResult | string, compact = false): TemplateResult {
  const cls = compact ? "sum-field-control sum-field-control--compact" : "sum-field-control";
  return html`<div class=${cls}>${body}</div>`;
}

export function fieldPlaceholder(field: SwcArchField): string {
  return field.placeholder ?? field.string ?? field.name;
}

export function fieldReadonlyValue(val: string, placeholder = ""): TemplateResult {
  const hasValue = val.trim() !== "";
  const text = hasValue ? val : placeholder;
  const cls = hasValue ? "sum-field-value" : "sum-field-value sum-field-value--placeholder";
  return html`<div class=${cls}>${text}</div>`;
}

export function fieldReadonlyInput(
  field: SwcArchField,
  val: string,
  inputType = "text",
): TemplateResult {
  const placeholder = fieldPlaceholder(field);
  return html`<input
    type=${inputType}
    class="sum-field-input"
    name=${field.name}
    value=${val}
    placeholder=${placeholder}
    readonly
    tabindex="-1"
  />`;
}

export interface FieldShellOptions {
  showLabel?: boolean;
  modifiers?: string[];
  labelFor?: string;
  layout?: "row" | "stack";
  compact?: boolean;
}

export function renderFieldShell(
  field: SwcArchField,
  body: TemplateResult | string,
  options: FieldShellOptions = {},
): TemplateResult {
  const showLabel = options.showLabel !== false;
  const labelFor = options.labelFor ?? fieldInputId(field);
  const useRow =
    options.layout === "row" || (options.layout !== "stack" && !isFullWidthField(field) && !options.compact);
  const modifiers = [...(options.modifiers ?? [])];
  if (useRow) modifiers.push("sum-field-widget--row");
  const wrappedBody = fieldControl(body, options.compact === true);

  if (useRow) {
    return html`<div class=${fieldWidgetClass(field, modifiers)}>
      ${showLabel ? fieldLabel(field, labelFor, true) : ""}
      ${wrappedBody}
    </div>`;
  }

  return html`<div class=${fieldWidgetClass(field, modifiers)}>
    ${showLabel ? fieldLabel(field, labelFor) : ""}
    ${wrappedBody}
  </div>`;
}
