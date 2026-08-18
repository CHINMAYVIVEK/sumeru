import { html, type TemplateResult } from "../core/template.js";
import type { SwcArchField } from "../types/workspace.js";

export function fieldInputId(field: SwcArchField): string {
  return `f-${field.name}`;
}

export function fieldWidgetClass(field: SwcArchField, extra: string[] = []): string {
  const parts = ["sum-field-widget"];
  if (field.type === "text" || field.widget === "text") {
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

export function fieldLabel(field: SwcArchField, forId?: string): TemplateResult {
  const label = field.string ?? field.name;
  if (forId) {
    return html`<label class="sum-field-label" for=${forId}>${label}</label>`;
  }
  return html`<label class="sum-field-label">${label}</label>`;
}

export function fieldReadonlyValue(val: string): TemplateResult {
  return html`<div class="sum-field-value">${val}</div>`;
}

export function fieldReadonlyInput(
  field: SwcArchField,
  val: string,
  inputType = "text",
): TemplateResult {
  return html`<input
    type=${inputType}
    class="sum-field-input"
    name=${field.name}
    value=${val}
    readonly
    tabindex="-1"
  />`;
}

export function renderFieldShell(
  field: SwcArchField,
  body: TemplateResult | string,
  options: { showLabel?: boolean; modifiers?: string[]; labelFor?: string } = {},
): TemplateResult {
  const showLabel = options.showLabel !== false;
  const labelFor = options.labelFor ?? fieldInputId(field);
  return html`<div class=${fieldWidgetClass(field, options.modifiers ?? [])}>
    ${showLabel ? fieldLabel(field, labelFor) : ""}
    ${body}
  </div>`;
}
