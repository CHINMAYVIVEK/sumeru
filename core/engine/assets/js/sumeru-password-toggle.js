"use strict";
(() => {
  // src/login/password-toggle.ts
  var EYE_OPEN_SVG = '<svg class="sum-password-icon sum-password-icon--show" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M2.062 12.348a1 1 0 0 1 0-.696a10.75 10.75 0 0 1 19.876 0a1 1 0 0 1 0 .696a10.75 10.75 0 0 1-19.876 0"/><circle cx="12" cy="12" r="3"/></svg>';
  var EYE_CLOSED_SVG = '<svg class="sum-password-icon sum-password-icon--hide" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10.733 5.076a10.744 10.744 0 0 1 11.205 6.575a1 1 0 0 1 0 .696a10.8 10.8 0 0 1-1.444 2.49m-6.41-.679a3 3 0 0 1-4.242-4.242"/><path d="M17.479 17.499a10.75 10.75 0 0 1-15.417-5.151a1 1 0 0 1 0-.696a10.75 10.75 0 0 1 4.446-5.143M2 2l20 20"/></svg>';
  function setVisible(input, btn, show) {
    input.type = show ? "text" : "password";
    btn.setAttribute("aria-pressed", show ? "true" : "false");
    btn.setAttribute("aria-label", show ? "Hide password" : "Show password");
    btn.classList.toggle("sum-password-toggle--revealed", show);
  }
  function enhanceInput(input) {
    if (input.closest(".sum-password-field")) {
      return;
    }
    const parent = input.parentNode;
    if (!parent) {
      return;
    }
    const wrapper = document.createElement("div");
    wrapper.className = "sum-password-field";
    parent.insertBefore(wrapper, input);
    wrapper.appendChild(input);
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "sum-password-toggle";
    btn.setAttribute("aria-label", "Show password");
    btn.setAttribute("aria-pressed", "false");
    btn.innerHTML = EYE_OPEN_SVG + EYE_CLOSED_SVG;
    wrapper.appendChild(btn);
    btn.addEventListener("click", () => {
      setVisible(input, btn, input.type === "password");
    });
  }
  function initPasswordToggles(root = document) {
    root.querySelectorAll('input[type="password"]').forEach(enhanceInput);
  }
  window.sumInitPasswordToggles = initPasswordToggles;
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => {
      initPasswordToggles(document);
    });
  } else {
    initPasswordToggles(document);
  }
})();
//# sourceMappingURL=sumeru-password-toggle.js.map
