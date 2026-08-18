import { beforeEach, describe, expect, it } from "vitest";
import { applyActivityHidden } from "../../src/shell/activity-panel.js";

describe("activity-panel", () => {
  beforeEach(() => {
    document.body.innerHTML = `
      <div id="sum-shell" class="sum-shell">
        <aside id="sum-activity-panel" class="sum-activity-panel"></aside>
        <button type="button" id="sum-activity-reveal" hidden></button>
        <button type="button" id="sum-activity-toggle" aria-pressed="true"></button>
      </div>
    `;
  });

  it("applyActivityHidden toggles shell class and reveal button", () => {
    const shell = document.getElementById("sum-shell") as HTMLElement;
    applyActivityHidden(shell, true);
    expect(shell.classList.contains("sum-shell--activity-hidden")).toBe(true);
    expect((document.getElementById("sum-activity-reveal") as HTMLButtonElement).hidden).toBe(false);

    applyActivityHidden(shell, false);
    expect(shell.classList.contains("sum-shell--activity-hidden")).toBe(false);
    expect((document.getElementById("sum-activity-reveal") as HTMLButtonElement).hidden).toBe(true);
  });
});
