/** Left sidebar: collapse, expand, persisted state. */

import { KEY_SIDEBAR, readBool, writeBool } from "./storage.js";
import { CHEVRON_RIGHT } from "../lib/icons.js";

function applySidebar(shell, collapsed) {
  shell.classList.toggle("sum-shell--sidebar-collapsed", collapsed);
  const reveal = document.getElementById("sum-sidebar-reveal");
  if (reveal) reveal.hidden = !collapsed;
}

function paintSidebarRevealIcon() {
  const sr = document.getElementById("sum-sidebar-reveal");
  if (sr && !sr.firstChild) sr.innerHTML = CHEVRON_RIGHT;
}

/**
 * @param {HTMLElement} shell
 */
export function initSidebar(shell) {
  applySidebar(shell, readBool(KEY_SIDEBAR));
  paintSidebarRevealIcon();

  function toggleSidebar() {
    const next = !shell.classList.contains("sum-shell--sidebar-collapsed");
    applySidebar(shell, next);
    writeBool(KEY_SIDEBAR, next);
  }

  ["sum-sidebar-toggle", "sum-sidebar-toggle-breadcrumb"].forEach((id) => {
    const el = document.getElementById(id);
    if (el) el.addEventListener("click", toggleSidebar);
  });

  const revealSidebar = document.getElementById("sum-sidebar-reveal");
  if (revealSidebar) {
    revealSidebar.addEventListener("click", () => {
      applySidebar(shell, false);
      writeBool(KEY_SIDEBAR, false);
    });
  }
}
