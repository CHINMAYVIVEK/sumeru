import { describe, expect, it } from "vitest";
import { forEach, when } from "../../src/template/helpers.js";
import { html } from "../../src/template/html.js";

describe("template helpers", () => {
  it("forEach assigns swcKey on rendered nodes", () => {
    const items = [{ id: 1, name: "A" }, { id: 2, name: "B" }];
    const results = forEach(items, (i) => i.id, (i) => html`<li>${i.name}</li>`);
    const wrap = document.createElement("ul");
    for (const r of results) wrap.appendChild(r.render());
    const keys = [...wrap.querySelectorAll("li")].map((el) => el.dataset.swcKey);
    expect(keys).toEqual(["1", "2"]);
  });

  it("when renders first matching branch", () => {
    const result = when(false, () => html`<span>a</span>`, [true, () => html`<span>b</span>`]);
    expect(result && typeof result === "object" && "render" in result ? result.render().textContent : "").toBe("b");
  });
});
