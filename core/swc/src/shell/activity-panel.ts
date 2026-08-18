import {
  KEY_ACTIVITY_HIDDEN,
  readActivityWidth,
  readBool,
  writeActivityWidth,
  writeBool,
} from "../util/shell-storage.js";

const CHEVRON_LEFT =
  '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M15 18l-6-6 6-6"/></svg>';

export function applyActivityWidth(px: number): void {
  document.documentElement.style.setProperty("--sum-activity-width", `${px}px`);
}

export function applyActivityHidden(shell: HTMLElement, hidden: boolean): void {
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

function paintActivityRevealIcon(): void {
  const reveal = document.getElementById("sum-activity-reveal");
  if (reveal && !reveal.firstChild) reveal.innerHTML = CHEVRON_LEFT;
}

function initActivityTabs(): void {
  document.querySelectorAll("[data-sum-activity-tab]").forEach((tab) => {
    tab.addEventListener("click", () => {
      const name = tab.getAttribute("data-sum-activity-tab");
      const panes: Record<string, string> = {
        messages: "sum-activity-pane-messages",
        log: "sum-activity-pane-log",
      };
      document.querySelectorAll("[data-sum-activity-tab]").forEach((t) => {
        const on = t.getAttribute("data-sum-activity-tab") === name;
        t.classList.toggle("is-active", on);
        t.setAttribute("aria-selected", on ? "true" : "false");
      });
      for (const [key, id] of Object.entries(panes)) {
        const el = document.getElementById(id);
        if (el) el.hidden = key !== name;
      }
    });
  });
}

function initActivityResizer(shell: HTMLElement): void {
  const resizer = document.getElementById("sum-activity-resizer");
  if (!resizer) return;

  let dragging = false;
  let startX = 0;
  let startW = 300;

  resizer.addEventListener("mousedown", (ev) => {
    if (shell.classList.contains("sum-shell--activity-hidden")) return;
    dragging = true;
    startX = ev.clientX;
    startW = readActivityWidth();
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
    ev.preventDefault();
  });

  window.addEventListener("mousemove", (ev) => {
    if (!dragging) return;
    const delta = startX - ev.clientX;
    let width = startW + delta;
    width = Math.min(520, Math.max(200, width));
    applyActivityWidth(width);
  });

  window.addEventListener("mouseup", () => {
    if (!dragging) return;
    dragging = false;
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
    const raw = document.documentElement.style.getPropertyValue("--sum-activity-width");
    const px = parseInt(raw, 10);
    if (!Number.isNaN(px)) writeActivityWidth(px);
  });
}

/** Wire the server-rendered activity dock in base.html. */
export function initActivityPanel(shell: HTMLElement): void {
  applyActivityWidth(readActivityWidth());
  applyActivityHidden(shell, readBool(KEY_ACTIVITY_HIDDEN));
  paintActivityRevealIcon();

  const setActivityHidden = (hidden: boolean): void => {
    applyActivityHidden(shell, hidden);
    writeBool(KEY_ACTIVITY_HIDDEN, hidden);
  };

  document.getElementById("sum-activity-toggle")?.addEventListener("click", () => {
    setActivityHidden(!shell.classList.contains("sum-shell--activity-hidden"));
  });

  document.getElementById("sum-activity-collapse")?.addEventListener("click", () => {
    setActivityHidden(true);
  });

  document.getElementById("sum-activity-reveal")?.addEventListener("click", () => {
    setActivityHidden(false);
  });

  initActivityTabs();
  initActivityResizer(shell);
}
