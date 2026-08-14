/**
 * Apps catalog: open module detail when clicking card/row chrome (not actions).
 * Loaded only on /web/apps via apps_inner.html.
 */

function isAppsPage() {
  return location.pathname === "/web/apps";
}

function openModuleDetail(mod, layout) {
  const q = new URLSearchParams();
  q.set("module", mod);
  q.set("layout", layout);
  const cur = new URLSearchParams(window.location.search);
  for (const key of ["filter", "scope", "q"]) {
    if (cur.has(key)) q.set(key, cur.get(key));
  }
  window.location.href = "/web/apps?" + q.toString();
}

export function initAppsModuleOpener() {
  if (!isAppsPage()) return;

  document.addEventListener(
    "click",
    (ev) => {
      const card = ev.target.closest(".sum-app-card-prem--open[data-module]");
      const row = ev.target.closest("tr.sum-apps-list-row--open[data-module]");
      const hit = card || row;
      if (!hit) return;
      if (ev.target.closest("form,button,a,.sum-app-card-prem-actions,.sum-apps-list-actions")) return;
      const mod = hit.getAttribute("data-module");
      const layout = hit.getAttribute("data-apps-layout") || "grid";
      if (!mod) return;
      openModuleDetail(mod, layout);
    },
    true
  );
}
