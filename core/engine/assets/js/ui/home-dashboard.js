/**
 * Home dashboard (/web/home): search shortcut, recent apps, pin to top bar.
 */
import { getPinnedApps, getRecentApps, togglePinnedApp, applyTopNavFilter } from "../core.js";

function loadLauncherApps() {
  const el = document.getElementById("sum-app-launcher-data");
  if (!el || !el.textContent) return [];
  try {
    return JSON.parse(el.textContent);
  } catch (_) {
    return [];
  }
}

function tileFromModule(mod, apps, card) {
  const app = apps.find((a) => a.name === mod);
  if (app) {
    return {
      href: app.href,
      displayName: app.displayName || mod,
      name: mod,
      letter: app.iconLetter || mod.charAt(0).toUpperCase(),
      hue: hashHue(mod),
    };
  }
  if (card) {
    const link = card.querySelector(".sum-home-hub-app-main");
    const letter = card.querySelector(".sum-home-hub-app-letter");
    return {
      href: link ? link.getAttribute("href") : "/web/home",
      displayName: card.querySelector(".sum-home-hub-app-name")?.textContent || mod,
      name: mod,
      letter: letter ? letter.textContent : mod.charAt(0).toUpperCase(),
      hue: hashHue(mod),
    };
  }
  return null;
}

function hashHue(s) {
  let h = 265;
  for (const c of String(s)) {
    h = (h * 31 + c.charCodeAt(0)) % 360;
  }
  return h < 0 ? h + 360 : h;
}

function renderRecentApps() {
  const section = document.getElementById("sum-home-recent");
  const container = document.getElementById("sum-home-recent-apps");
  if (!section || !container) return;

  const recents = getRecentApps();
  if (recents.length === 0) {
    section.hidden = true;
    return;
  }

  const apps = loadLauncherApps();
  const cardsByMod = {};
  document.querySelectorAll(".sum-home-hub-app[data-module]").forEach((card) => {
    cardsByMod[card.getAttribute("data-module")] = card;
  });

  container.innerHTML = "";
  recents.forEach((mod) => {
    const t = tileFromModule(mod, apps, cardsByMod[mod]);
    if (!t) return;
    const item = document.createElement("div");
    item.className = "sum-home-hub-app";
    item.setAttribute("role", "listitem");
    item.setAttribute("data-module", mod);
    item.innerHTML = `<a href="${t.href}" class="sum-home-hub-app-main" data-module="${mod}" title="${t.displayName}">
      <div class="sum-home-hub-app-icon" style="--sum-home-icon-h: ${t.hue};" aria-hidden="true">
        <span class="sum-home-hub-app-letter">${t.letter}</span>
      </div>
      <span class="sum-home-hub-app-name">${t.displayName}</span>
      <span class="sum-home-hub-app-tech">${t.name}</span>
    </a>`;
    container.appendChild(item);
  });

  section.hidden = container.children.length === 0;
}

function paintPinButtons() {
  const pins = getPinnedApps();
  document.querySelectorAll(".sum-home-hub-app-pin").forEach((btn) => {
    const mod = btn.getAttribute("data-module");
    btn.classList.toggle("is-pinned", pins.includes(mod));
  });

  const grid = document.querySelector(".sum-home-hub-apps:not(#sum-home-recent-apps)");
  if (!grid || pins.length === 0) return;
  const items = [...grid.querySelectorAll(".sum-home-hub-app[data-module]")];
  items.sort((a, b) => {
    const am = a.getAttribute("data-module");
    const bm = b.getAttribute("data-module");
    const ap = pins.includes(am);
    const bp = pins.includes(bm);
    if (ap !== bp) return ap ? -1 : 1;
    return 0;
  });
  items.forEach((el) => grid.appendChild(el));
}

(function initHomeDashboard() {
  const input = document.getElementById("sum-home-q");
  if (input) {
    document.addEventListener("keydown", (e) => {
      if (e.key !== "/" || e.ctrlKey || e.metaKey || e.altKey) return;
      const tag = document.activeElement && document.activeElement.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if (document.getElementById("sum-app-launcher")?.open) return;
      e.preventDefault();
      input.focus();
      try {
        input.select();
      } catch (_) {}
    });
  }

  document.querySelectorAll(".sum-home-hub-app-pin").forEach((btn) => {
    btn.addEventListener("click", (e) => {
      e.preventDefault();
      e.stopPropagation();
      togglePinnedApp(btn.getAttribute("data-module"));
      paintPinButtons();
      applyTopNavFilter();
    });
  });

  renderRecentApps();
  paintPinButtons();
})();
