/** Right activity dock: width, hide/show, Messages/Log tabs, resize handle. */

import {
  KEY_ACTIVITY_HIDDEN,
  readBool,
  writeBool,
  readActivityWidth,
  writeActivityWidth,
} from "./storage.js";
import { CHEVRON_LEFT } from "../lib/icons.js";

function applyActivityWidth(px) {
  document.documentElement.style.setProperty("--sum-activity-width", px + "px");
}

function applyActivityHidden(shell, hidden) {
  shell.classList.toggle("sum-shell--activity-hidden", hidden);
  const reveal = document.getElementById("sum-activity-reveal");
  if (reveal) reveal.hidden = !hidden;
  const toggle = document.getElementById("sum-activity-toggle");
  if (toggle) {
    const pressed = !hidden;
    toggle.setAttribute("aria-pressed", pressed ? "true" : "false");
    toggle.classList.toggle("is-pressed", pressed);
  }
}

function paintActivityRevealIcon() {
  const ar = document.getElementById("sum-activity-reveal");
  if (ar && !ar.firstChild) ar.innerHTML = CHEVRON_LEFT;
}

function initActivityTabs() {
  document.querySelectorAll("[data-sum-activity-tab]").forEach((tab) => {
    tab.addEventListener("click", () => {
      const name = tab.getAttribute("data-sum-activity-tab");
      const panes = { messages: "sum-activity-pane-messages", log: "sum-activity-pane-log" };
      document.querySelectorAll("[data-sum-activity-tab]").forEach((t) => {
        const on = t.getAttribute("data-sum-activity-tab") === name;
        t.classList.toggle("is-active", on);
        t.setAttribute("aria-selected", on ? "true" : "false");
      });
      Object.entries(panes).forEach(([k, id]) => {
        const el = document.getElementById(id);
        if (!el) return;
        el.hidden = k !== name;
      });
    });
  });
}

function initActivityResizer(shell) {
  const resizer = document.getElementById("sum-activity-resizer");
  if (!resizer) return;

  let dragging = false;
  let startX = 0;
  let startW = 300;

  resizer.addEventListener("mousedown", (e) => {
    if (shell.classList.contains("sum-shell--activity-hidden")) return;
    dragging = true;
    startX = e.clientX;
    startW = readActivityWidth();
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
    e.preventDefault();
  });

  window.addEventListener("mousemove", (e) => {
    if (!dragging) return;
    const delta = startX - e.clientX;
    let w = startW + delta;
    w = Math.min(520, Math.max(200, w));
    applyActivityWidth(w);
  });

  window.addEventListener("mouseup", () => {
    if (!dragging) return;
    dragging = false;
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
    const m = document.documentElement.style.getPropertyValue("--sum-activity-width");
    const px = parseInt(m, 10);
    if (!Number.isNaN(px)) writeActivityWidth(px);
  });
}

/**
 * @param {HTMLElement} shell
 */
export function initActivityPanel(shell) {
  applyActivityWidth(readActivityWidth());
  applyActivityHidden(shell, readBool(KEY_ACTIVITY_HIDDEN));
  paintActivityRevealIcon();

  function setActivityHidden(hidden) {
    applyActivityHidden(shell, hidden);
    writeBool(KEY_ACTIVITY_HIDDEN, hidden);
  }

  const activityToggle = document.getElementById("sum-activity-toggle");
  if (activityToggle) {
    activityToggle.addEventListener("click", () => {
      setActivityHidden(!shell.classList.contains("sum-shell--activity-hidden"));
    });
  }

  const activityCollapse = document.getElementById("sum-activity-collapse");
  if (activityCollapse) {
    activityCollapse.addEventListener("click", () => setActivityHidden(true));
  }

  const activityReveal = document.getElementById("sum-activity-reveal");
  if (activityReveal) {
    activityReveal.addEventListener("click", () => setActivityHidden(false));
  }

  initActivityTabs();
  initActivityResizer(shell);
}
