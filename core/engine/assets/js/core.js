/**
 * Shell bootstrap: wires feature modules (sidebar, activity, forms, apps).
 * Each imported module owns one concern.
 */
import { initSidebar } from "./shell/sidebar.js";
import { initTopbarDropdowns } from "./shell/topbar-dropdowns.js";
import { initActivityPanel } from "./shell/activity-panel.js";
import { initNotebookTabs } from "./ui/notebook-tabs.js";
import { initAjaxFormCapture } from "./ui/ajax-form.js";
import { initAppsModuleOpener } from "./apps/apps-module-opener.js";
import { initFormSplit } from "./ui/form-split.js";
import { initMessagesComposer } from "./ui/messages-composer.js";
import { initMany2One } from "./ui/many2one.js";
import { initMany2OneSelect } from "./ui/many2one-select.js";
import { initMultiSelect } from "./ui/multi-select.js";
import { initAvatarUpload } from "./ui/avatar-upload.js";
import { initKanbanBoard } from "./ui/kanban-board.js";
import { initStatusbar } from "./ui/statusbar.js";

const KEY_PINNED = "sumeru:pinned-apps";
const KEY_RECENT = "sumeru:recent-apps";
const MAX_RECENT = 5;
const MAX_TOPBAR_RECENT = 4;

function readJSON(key, fallback) {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return fallback;
    const v = JSON.parse(raw);
    return Array.isArray(v) ? v : fallback;
  } catch (_) {
    return fallback;
  }
}

function writeJSON(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch (_) {}
}

export function getPinnedApps() {
  return readJSON(KEY_PINNED, []);
}

export function getRecentApps() {
  return readJSON(KEY_RECENT, []);
}

export function togglePinnedApp(moduleName) {
  const mod = String(moduleName || "").trim();
  if (!mod) return getPinnedApps();
  let pins = getPinnedApps();
  if (pins.includes(mod)) {
    pins = pins.filter((m) => m !== mod);
  } else {
    pins = [mod, ...pins];
  }
  writeJSON(KEY_PINNED, pins);
  return pins;
}

export function pushRecentApp(moduleName) {
  const mod = String(moduleName || "").trim();
  if (!mod) return;
  let recents = getRecentApps().filter((m) => m !== mod);
  recents.unshift(mod);
  if (recents.length > MAX_RECENT) recents = recents.slice(0, MAX_RECENT);
  writeJSON(KEY_RECENT, recents);
}

function loadLauncherApps() {
  const el = document.getElementById("sum-app-launcher-data");
  if (!el || !el.textContent) return [];
  try {
    return JSON.parse(el.textContent);
  } catch (_) {
    return [];
  }
}

function launcherMatches(q, item) {
  const hay = `${item.displayName || ""} ${item.name || ""} ${item.description || ""}`.toLowerCase();
  const tokens = String(q || "")
    .trim()
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean);
  if (tokens.length === 0) return true;
  return tokens.every((t) => hay.includes(t));
}

export function applyTopNavFilter() {
  const nav = document.querySelector(".sum-top-nav");
  if (!nav) return;

  const moduleItems = [...nav.querySelectorAll(".top-menu-item--module")];
  if (moduleItems.length === 0) return;

  const pins = getPinnedApps();
  const recents = getRecentApps();
  const active = nav.querySelector(".top-menu-item--module.active");
  const activeMod = active ? active.getAttribute("data-module") : "";

  let visibleMods = new Set();
  if (pins.length > 0) {
    pins.forEach((m) => visibleMods.add(m));
  } else if (recents.length > 0) {
    recents.slice(0, MAX_TOPBAR_RECENT).forEach((m) => visibleMods.add(m));
  } else {
    moduleItems.forEach((el) => visibleMods.add(el.getAttribute("data-module") || ""));
  }
  if (activeMod) visibleMods.add(activeMod);

  const shouldFilter = pins.length > 0 || recents.length > 0;
  if (!shouldFilter) return;

  moduleItems.forEach((el) => {
    const mod = el.getAttribute("data-module") || "";
    el.classList.toggle("is-topbar-hidden", !visibleMods.has(mod));
  });

  const activeEl = nav.querySelector(".top-menu-item.active");
  if (activeEl && typeof activeEl.scrollIntoView === "function") {
    activeEl.scrollIntoView({ inline: "nearest", block: "nearest", behavior: "instant" });
  }
}

function initRecentTracking() {
  document.addEventListener(
    "click",
    (e) => {
      const link = e.target.closest("[data-module]");
      if (!link) return;
      const mod = link.getAttribute("data-module");
      if (mod) pushRecentApp(mod);
    },
    true
  );
}

function initAppLauncher() {
  const dialog = document.getElementById("sum-app-launcher");
  const input = document.getElementById("sum-app-launcher-input");
  const resultsEl = document.getElementById("sum-app-launcher-results");
  if (!dialog || !input || !resultsEl) return;

  const apps = loadLauncherApps();
  let activeIdx = 0;
  let filtered = apps.slice();

  function renderResults() {
    resultsEl.innerHTML = "";
    if (filtered.length === 0) {
      const li = document.createElement("li");
      li.className = "sum-app-launcher-result";
      li.textContent = "No applications match.";
      li.setAttribute("role", "option");
      resultsEl.appendChild(li);
      return;
    }
    filtered.forEach((item, i) => {
      const li = document.createElement("li");
      li.className = "sum-app-launcher-result" + (i === activeIdx ? " is-active" : "");
      li.setAttribute("role", "option");
      li.dataset.href = item.href || "/web/home";
      li.dataset.module = item.name || "";
      li.innerHTML = `<span class="sum-app-launcher-result-letter" aria-hidden="true">${item.iconLetter || "?"}</span>
        <span class="sum-app-launcher-result-body">
          <div class="sum-app-launcher-result-name">${item.displayName || item.name}</div>
          <div class="sum-app-launcher-result-meta">${item.name || ""}${item.description ? " · " + item.description : ""}</div>
        </span>`;
      li.addEventListener("mousedown", (ev) => {
        ev.preventDefault();
        openLauncherItem(item);
      });
      resultsEl.appendChild(li);
    });
  }

  function openLauncherItem(item) {
    if (!item || !item.href) return;
    if (item.name) pushRecentApp(item.name);
    close();
    window.location.href = item.href;
  }

  function filter(q) {
    filtered = apps.filter((item) => launcherMatches(q, item));
    activeIdx = 0;
    renderResults();
  }

  function open() {
    if (typeof dialog.showModal === "function") {
      dialog.showModal();
    } else {
      dialog.setAttribute("open", "");
    }
    input.value = "";
    filter("");
    requestAnimationFrame(() => input.focus());
  }

  function close() {
    if (dialog.open) dialog.close();
    dialog.removeAttribute("open");
  }

  input.addEventListener("input", () => filter(input.value));

  input.addEventListener("keydown", (e) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      activeIdx = Math.min(activeIdx + 1, Math.max(0, filtered.length - 1));
      renderResults();
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      activeIdx = Math.max(activeIdx - 1, 0);
      renderResults();
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (filtered[activeIdx]) openLauncherItem(filtered[activeIdx]);
    } else if (e.key === "Escape") {
      e.preventDefault();
      close();
    }
  });

  dialog.addEventListener("click", (e) => {
    if (e.target === dialog) close();
  });

  document.addEventListener("keydown", (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
      e.preventDefault();
      if (dialog.open) close();
      else open();
    }
  });
}

(function bootstrap() {
  const shell = document.getElementById("sum-shell");
  if (!shell) return;

  initSidebar(shell);
  initTopbarDropdowns(shell);
  initActivityPanel(shell);
  initNotebookTabs();
  initAjaxFormCapture();
  initAppsModuleOpener();
  initFormSplit();
  initMessagesComposer();
  initMany2One();
  initMany2OneSelect();
  initMultiSelect();
  initAvatarUpload();
  initKanbanBoard();
  initStatusbar();
  initRecentTracking();
  applyTopNavFilter();
  initAppLauncher();
})();
