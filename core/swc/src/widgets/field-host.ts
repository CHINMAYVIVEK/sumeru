import type { SwcEnv } from "../runtime/env.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import { registry } from "../runtime/registry.js";
import { resolveFieldWidget } from "./registry.js";
import { DefaultField } from "./DefaultField.js";

interface FieldEntry {
  comp: {
    render(): HTMLElement;
    destroy(): void;
    setup?(): void;
  };
  readonly: boolean;
  widget: string;
}

/** Reuses field widget instances across FormView patches (same record + mode). */
export class FieldHost {
  private readonly env: SwcEnv;
  private readonly entries = new Map<string, FieldEntry>();

  constructor(env: SwcEnv) {
    this.env = env;
  }

  render(field: SwcArchField, record: SwcRecord, readonly: boolean): HTMLElement {
    const widget = resolveFieldWidget(field);
    const key = field.name;
    const prev = this.entries.get(key);

    if (prev && prev.readonly === readonly && prev.widget === widget) {
      return prev.comp.render();
    }

    prev?.comp.destroy();
    const Ctor = (registry.get("fields", widget) ?? registry.get("fields", "default")) as unknown as typeof DefaultField;
    const comp = new Ctor({ field, record, readonly }, this.env);
    comp.setup?.();
    this.entries.set(key, { comp, readonly, widget });
    return comp.render();
  }

  clear(): void {
    for (const { comp } of this.entries.values()) {
      comp.destroy();
    }
    this.entries.clear();
  }
}
