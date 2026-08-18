import { describe, expect, it, vi } from "vitest";
import { SwcRecord } from "../../src/model/record.js";
import { registerDefaultWidgets, renderField } from "../../src/widgets/registry.js";
import type { SwcArchField } from "../../src/types/workspace.js";
import type { SwcEnv } from "../../src/runtime/env.js";

registerDefaultWidgets();

const env = {
  bootstrap: {} as never,
  services: {
    rpc: {
      searchRead: vi.fn().mockResolvedValue([]),
    },
  },
} as unknown as SwcEnv;

function field(overrides: Partial<SwcArchField>): SwcArchField {
  return { name: "x", ...overrides };
}

describe("form field widgets", () => {
  it("default char field emits sum-field-widget shell", () => {
    const record = new SwcRecord("m", 1, { name: "Acme" });
    const el = renderField(env, field({ name: "name", string: "Name", type: "char" }), record, false);
    expect(el.classList.contains("sum-field-widget")).toBe(true);
    expect(el.querySelector(".sum-field-input")).toBeTruthy();
  });

  it("phone widget uses tel input", () => {
    const record = new SwcRecord("m", 1, { phone: "555" });
    const el = renderField(env, field({ name: "phone", widget: "phone" }), record, false);
    expect(el.querySelector('input[type="tel"]')).toBeTruthy();
  });

  it("date field uses date picker popover", () => {
    const record = new SwcRecord("m", 1, { date_start: "2026-01-01" });
    const el = renderField(env, field({ name: "date_start", type: "date", string: "Start Date" }), record, false);
    expect(el.querySelector(".sum-date-field")).toBeTruthy();
    expect(el.querySelector(".sum-date-popover-input")?.getAttribute("type")?.trim()).toBe("date");
  });

  it("many2many_tags renders tag container", () => {
    const record = new SwcRecord("m", 1, { tag_ids: [1] });
    const el = renderField(
      env,
      field({ name: "tag_ids", widget: "many2many_tags", relation: "my.module.tag" }),
      record,
      true,
    );
    expect(el.querySelector(".sum-field-tags, .sum-multi-select-tags")).toBeTruthy();
  });

  it("radio widget renders radio group", () => {
    const record = new SwcRecord("m", 1, { active: true });
    const el = renderField(env, field({ name: "active", type: "boolean", widget: "radio" }), record, false);
    expect(el.querySelector(".sum-field-radio-group")).toBeTruthy();
  });
});
