import { registry, type RegistryEntry } from "../core/registry.js";
import type { SwcArchField } from "../types/workspace.js";
import { DefaultField } from "./DefaultField.js";
import { Many2OneField } from "./Many2OneField.js";
import { StatusbarField } from "./StatusbarField.js";
import { PriorityField } from "./PriorityField.js";
import { BooleanField } from "./BooleanField.js";
import { TextareaField } from "./TextareaField.js";
import { SelectionField } from "./SelectionField.js";
import { PhoneField } from "./PhoneField.js";
import { BooleanRadioField } from "./BooleanRadioField.js";
import { BooleanToggleField } from "./BooleanToggleField.js";
import { Many2ManyTagsField } from "./Many2ManyTagsField.js";
import { One2ManyField } from "./One2ManyField.js";
import { ImageField } from "./ImageField.js";

export function registerDefaultWidgets(): void {
  const fields = registry.category("fields");
  const add = (key: string, Ctor: unknown) => fields.add(key, Ctor as RegistryEntry);

  add("default", DefaultField);
  add("char", DefaultField);
  add("email", DefaultField);
  add("integer", DefaultField);
  add("float", DefaultField);
  add("numeric", DefaultField);
  add("date", DefaultField);
  add("datetime", DefaultField);
  add("json", TextareaField);
  add("many2one", Many2OneField);
  add("one2many", One2ManyField);
  add("many2many", Many2ManyTagsField);
  add("selection", SelectionField);
  add("boolean", BooleanField);
  add("text", TextareaField);
  add("statusbar", StatusbarField);
  add("priority", PriorityField);
  add("phone", PhoneField);
  add("radio", BooleanRadioField);
  add("boolean_toggle", BooleanToggleField);
  add("many2many_tags", Many2ManyTagsField);
  add("image", ImageField);
}

export function resolveFieldWidget(field: SwcArchField): string {
  if (field.widget === "many2many_tags") return "many2many_tags";
  if (field.widget === "boolean_toggle") return "boolean_toggle";
  if (field.widget === "radio") return "radio";
  if (field.widget === "phone") return "phone";
  if (field.widget === "image") return "image";
  if (field.widget === "selection") return "selection";
  if (field.widget === "email") return "email";
  if (field.widget === "statusbar") return "statusbar";
  if (field.widget === "priority") return "priority";
  if (field.type === "boolean" && field.widget === "radio") return "radio";
  if (field.type === "boolean") return "boolean";
  if (field.type === "text") return "text";
  if (field.type === "many2one") return "many2one";
  if (field.type === "one2many") return "one2many";
  if (field.type === "many2many") return "many2many_tags";
  if (field.type === "selection") return "selection";
  if (field.type === "date") return "date";
  if (field.type === "datetime") return "datetime";
  if (field.type === "integer" || field.type === "float" || field.type === "numeric") {
    return field.type;
  }
  return field.widget ?? field.type ?? "default";
}

export function renderField(
  env: import("../core/env.js").SwcEnv,
  field: SwcArchField,
  record: import("../store/record.js").SwcRecord,
  readonly: boolean,
): HTMLElement {
  const key = resolveFieldWidget(field);
  const Ctor = (registry.get("fields", key) ?? registry.get("fields", "default")) as unknown as typeof DefaultField;
  const comp = new Ctor({ field, record, readonly }, env);
  comp.setup?.();
  return comp.render();
}

export class FieldRegistry {
  static render = renderField;
}
