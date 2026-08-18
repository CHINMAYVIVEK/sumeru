import { describe, expect, it } from "vitest";
import { SwcRecord } from "../src/store/record.js";
import { PriorityField } from "../src/widgets/PriorityField.js";
import type { SwcArchField } from "../src/types/workspace.js";
import type { SwcEnv } from "../src/core/env.js";

function field(overrides: Partial<SwcArchField> = {}): SwcArchField {
  return {
    name: "priority",
    string: "Priority",
    type: "selection",
    widget: "priority",
    selection: [
      ["0", "Low"],
      ["1", "Medium"],
      ["2", "High"],
      ["3", "Very High"],
    ],
    ...overrides,
  };
}

describe("PriorityField", () => {
  it("renders star buttons mapped to selection values", () => {
    const env = { bootstrap: {} as never, services: {} as never } as SwcEnv;
    const record = new SwcRecord("crm.lead", 1, { priority: "2" });
    const comp = new PriorityField({ field: field(), record, readonly: false }, env);
    const el = comp.render();
    expect(el.querySelectorAll(".sum-priority-star").length).toBe(3);
    expect(el.querySelectorAll(".sum-priority-star--on").length).toBe(2);
  });

  it("renders dropdown when mode is select", () => {
    const env = { bootstrap: {} as never, services: {} as never } as SwcEnv;
    const record = new SwcRecord("crm.lead", 1, { priority: "1" });
    const comp = new PriorityField(
      { field: field({ options: { mode: "select" } }), record, readonly: false },
      env,
    );
    const el = comp.render();
    expect(el.querySelector(".sum-priority-select")).toBeTruthy();
    expect(el.querySelector(".sum-priority-stars")).toBeNull();
  });
});
