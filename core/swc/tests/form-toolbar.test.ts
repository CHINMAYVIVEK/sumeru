import { describe, expect, it } from "vitest";
import { resolveHeaderButtonClass } from "../src/views/view-toolbar.js";

/** Mirrors FormView toolbar mode logic for unit testing. */
function toolbarMode(recordId: number, editing: boolean): "readonly" | "editing" {
  if (recordId > 0 && !editing) return "readonly";
  return "editing";
}

function toolbarLabels(recordId: number, editing: boolean): string[] {
  const mode = toolbarMode(recordId, editing);
  if (mode === "readonly") return ["New", "Edit"];
  return ["Save", "Cancel"];
}

describe("form toolbar modes", () => {
  it("existing readonly shows New and Edit", () => {
    expect(toolbarLabels(7, false)).toEqual(["New", "Edit"]);
  });

  it("existing editing shows Save and Cancel", () => {
    expect(toolbarLabels(7, true)).toEqual(["Save", "Cancel"]);
  });

  it("new record shows Save and Cancel", () => {
    expect(toolbarLabels(0, true)).toEqual(["Save", "Cancel"]);
  });
});

describe("header button class mapping", () => {
  it("maps sum_highlight to primary", () => {
    expect(resolveHeaderButtonClass("sum_highlight")).toContain("sum-header-btn--primary");
  });

  it("defaults to secondary", () => {
    expect(resolveHeaderButtonClass("")).toContain("sum-header-btn--secondary");
    expect(resolveHeaderButtonClass(undefined)).toContain("sum-header-btn--secondary");
  });
});
