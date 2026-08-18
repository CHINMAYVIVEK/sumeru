"use strict";
(() => {
  // devtools-extension/src/types.ts
  function parseDevtoolsSnapshot(raw) {
    if (typeof raw !== "string" || !raw) {
      return { components: [] };
    }
    try {
      const parsed = JSON.parse(raw);
      return { components: parsed.components ?? [] };
    } catch {
      return { components: [] };
    }
  }
  var DEVTOOLS_SNAPSHOT_EXPR = `window.__SWC_DEVTOOLS__
  ? JSON.stringify({
      components: window.__SWC_DEVTOOLS__.components.map((c) => ({
        id: c.id,
        name: c.name,
      })),
    })
  : ""`;

  // devtools-extension/src/panel.ts
  var statusEl = document.getElementById("status");
  var treeEl = document.getElementById("tree");
  var REFRESH_MS = 2e3;
  function renderTree(components) {
    if (!treeEl) return;
    if (components.length === 0) {
      treeEl.innerHTML = `<li class="sum-devtools-empty">No components</li>`;
      return;
    }
    treeEl.innerHTML = components.map((c) => `<li class="sum-devtools-node">${c.name} <span>#${c.id}</span></li>`).join("");
  }
  function refresh() {
    chrome.devtools.inspectedWindow.eval(DEVTOOLS_SNAPSHOT_EXPR, (result, err) => {
      if (err || !result) {
        if (statusEl) statusEl.textContent = "SWC not detected on this page.";
        renderTree([]);
        return;
      }
      const snapshot = parseDevtoolsSnapshot(result);
      if (statusEl) {
        statusEl.textContent = snapshot.components.length === 0 ? "SWC loaded \u2014 no components mounted yet." : `${snapshot.components.length} component(s)`;
      }
      renderTree(snapshot.components);
    });
  }
  refresh();
  window.setInterval(refresh, REFRESH_MS);
})();
//# sourceMappingURL=panel.js.map
