import { describe, it, expect } from "vitest";
import { parseFilterCSV, toggleFilterName } from "../../src/views/list/control-panel.js";

describe("search filter chips", () => {
  it("parses comma-separated filter names", () => {
    expect(parseFilterCSV("won, my")).toEqual(["won", "my"]);
    expect(parseFilterCSV("")).toEqual([]);
  });

  it("toggles filter names", () => {
    expect(toggleFilterName(["won"], "my")).toEqual(["won", "my"]);
    expect(toggleFilterName(["won", "my"], "won")).toEqual(["my"]);
  });
});
