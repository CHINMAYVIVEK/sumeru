import { describe, it, expect } from "vitest";
import { fieldModifiers, isFieldVisible } from "../../src/model/modifiers.js";

describe("field modifiers", () => {
  it("respects static invisible flag", () => {
    expect(isFieldVisible({ name: "x", invisible: true })).toBe(false);
    expect(isFieldVisible({ name: "x" })).toBe(true);
  });

  it("merges record overrides", () => {
    const field = { name: "x", readonly: false };
    const mods = fieldModifiers(field, {
      modifierOverrides: new Map([["x", { readonly: true }]]),
    } as never);
    expect(mods.readonly).toBe(true);
  });
});
