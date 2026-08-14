/** Cmd+K / topbar search launcher — DOM: `#sum-app-launcher`, `#sum-app-launcher-input`. */

import { loadLauncherApps } from "./launcher-data.js";
import { pushRecentApp } from "./pinned-apps.js";
import { tokenizeQuery, matchesAllTokens } from "../lib/token-match.js";

function launcherMatches(query, item) {
  const kindLabel = item.kind === "menu" ? "menu" : "app";
  const searchText = `${kindLabel} ${item.displayName || ""} ${item.name || ""} ${item.description || ""}`;
  return matchesAllTokens(searchText, tokenizeQuery(query));
}

export function initAppLauncher() {
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
      li.textContent = "No apps or menus match.";
      li.setAttribute("role", "option");
      resultsEl.appendChild(li);
      return;
    }
    filtered.forEach((item, i) => {
      const li = document.createElement("li");
      li.className = "sum-app-launcher-result" + (i === activeIdx ? " is-active" : "");
      li.setAttribute("role", "option");
      li.dataset.href = item.href || "/web/home";
      li.dataset.module = item.kind === "menu" ? "" : item.name || "";
      const kindPrefix = item.kind === "menu" ? "Menu · " : "App · ";
      const metaBody = item.kind === "menu"
        ? (item.description || "Menu")
        : `${item.name || ""}${item.description ? " · " + item.description : ""}`;
      li.innerHTML = `<span class="sum-app-launcher-result-letter" aria-hidden="true">${item.iconLetter || "?"}</span>
        <span class="sum-app-launcher-result-body">
          <div class="sum-app-launcher-result-name">${item.displayName || item.name}</div>
          <div class="sum-app-launcher-result-meta">${kindPrefix}${metaBody}</div>
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
    if (item.name && item.kind !== "menu") pushRecentApp(item.name);
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

  const topSearchOpen = document.getElementById("sum-topbar-search-open");
  const topSearchField = document.getElementById("sum-topbar-search-field");
  topSearchOpen?.addEventListener("click", (e) => {
    e.preventDefault();
    open();
  });
  topSearchField?.addEventListener("click", (e) => {
    e.preventDefault();
    open();
  });
  topSearchField?.addEventListener("focus", (e) => {
    e.preventDefault();
    topSearchField.blur();
    open();
  });
}
