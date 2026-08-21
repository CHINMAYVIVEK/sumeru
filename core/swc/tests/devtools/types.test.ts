import { describe, expect, it } from "vitest";
import {
  DEVTOOLS_SNAPSHOT_EXPR,
  parseDevtoolsSnapshot,
} from "../../devtools-extension/src/types.js";

describe("devtools extension types", () => {
  it("parses component snapshot JSON", () => {
    const raw = JSON.stringify({ components: [{ id: 1, name: "ListView" }] });
    expect(parseDevtoolsSnapshot(raw)).toEqual({
      components: [{ id: 1, name: "ListView" }],
    });
  });

  it("returns empty snapshot for missing SWC", () => {
    expect(parseDevtoolsSnapshot("")).toEqual({ components: [] });
    expect(parseDevtoolsSnapshot("not-json")).toEqual({ components: [] });
  });

  it("exports eval expression for inspected window", () => {
    expect(DEVTOOLS_SNAPSHOT_EXPR).toContain("__SWC_DEVTOOLS__");
  });
});
