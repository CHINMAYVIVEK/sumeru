import { describe, expect, it } from "vitest";
import { patchKeyedChildren } from "../../src/runtime/patch/keyed.js";

describe("keyed patch", () => {
  it("reorders and preserves nodes by key", () => {
    const tbody = document.createElement("tbody");
    const a = document.createElement("tr");
    a.dataset.swcKey = "1";
    a.textContent = "A";
    const b = document.createElement("tr");
    b.dataset.swcKey = "2";
    b.textContent = "B";
    tbody.append(a, b);

    patchKeyedChildren(tbody, [
      { key: "2", render: () => { const r = document.createElement("tr"); r.textContent = "B-new"; return r; } },
      { key: "1", render: () => { const r = document.createElement("tr"); r.textContent = "A-new"; return r; } },
    ]);

    expect([...tbody.children].map((c) => c.dataset.swcKey)).toEqual(["2", "1"]);
    expect(tbody.children[0].textContent).toBe("B");
    expect(tbody.children[1].textContent).toBe("A");
  });
});
