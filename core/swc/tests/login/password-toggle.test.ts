import { describe, it, expect, afterEach } from "vitest";
import { initPasswordToggles } from "../../src/login/password-toggle.js";

describe("initPasswordToggles", () => {
  afterEach(() => {
    document.body.innerHTML = "";
  });

  it("wraps password inputs with a toggle button", () => {
    document.body.innerHTML =
      '<div class="field"><input id="pw" type="password" value="secret" /></div>';
    initPasswordToggles(document);

    const input = document.getElementById("pw") as HTMLInputElement;
    const wrapper = input.closest(".sum-password-field");
    const btn = wrapper?.querySelector(".sum-password-toggle") as HTMLButtonElement;

    expect(wrapper).toBeTruthy();
    expect(btn).toBeTruthy();
    expect(btn.getAttribute("aria-pressed")).toBe("false");
    expect(wrapper?.querySelector(".sum-password-icon--show")).toBeTruthy();
    expect(wrapper?.querySelector(".sum-password-icon--hide")).toBeTruthy();
  });

  it("shows only one icon at a time when toggled", () => {
    document.body.innerHTML = '<input id="pw" type="password" />';
    initPasswordToggles(document);

    const input = document.getElementById("pw") as HTMLInputElement;
    const btn = document.querySelector(".sum-password-toggle") as HTMLButtonElement;

    btn.click();
    expect(input.type).toBe("text");
    expect(btn.classList.contains("sum-password-toggle--revealed")).toBe(true);
    expect(btn.getAttribute("aria-pressed")).toBe("true");

    btn.click();
    expect(input.type).toBe("password");
    expect(btn.classList.contains("sum-password-toggle--revealed")).toBe(false);
    expect(btn.getAttribute("aria-pressed")).toBe("false");
  });

  it("does not double-wrap inputs", () => {
    document.body.innerHTML = '<input type="password" /><input type="password" />';
    initPasswordToggles(document);
    initPasswordToggles(document);
    expect(document.querySelectorAll(".sum-password-field").length).toBe(2);
  });
});
