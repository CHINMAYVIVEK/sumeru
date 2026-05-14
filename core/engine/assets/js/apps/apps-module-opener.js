/**
 * Apps catalog: open module detail when clicking card/row chrome (not actions).
 * Only registered on /web/apps.
 */

export function initAppsModuleOpener() {
  if (location.pathname !== "/web/apps") return;

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
      const q = new URLSearchParams();
      q.set("module", mod);
      q.set("layout", layout);
      const cur = new URLSearchParams(window.location.search);
      for (const key of ["filter", "scope", "q"]) {
        if (cur.has(key)) q.set(key, cur.get(key));
      }
      window.location.href = "/web/apps?" + q.toString();
    },
    true
  );
}
