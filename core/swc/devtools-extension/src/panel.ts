import { DEVTOOLS_SNAPSHOT_EXPR, parseDevtoolsSnapshot } from "./types.js";

const statusEl = document.getElementById("status");
const treeEl = document.getElementById("tree");
const REFRESH_MS = 2000;

function renderTree(components: Array<{ id: number; name: string }>): void {
  if (!treeEl) return;
  if (components.length === 0) {
    treeEl.innerHTML = `<li class="sum-devtools-empty">No components</li>`;
    return;
  }
  treeEl.innerHTML = components
    .map((c) => `<li class="sum-devtools-node">${c.name} <span>#${c.id}</span></li>`)
    .join("");
}

function refresh(): void {
  chrome.devtools.inspectedWindow.eval<string>(DEVTOOLS_SNAPSHOT_EXPR, (result, err) => {
    if (err || !result) {
      if (statusEl) statusEl.textContent = "SWC not detected on this page.";
      renderTree([]);
      return;
    }
    const snapshot = parseDevtoolsSnapshot(result);
    if (statusEl) {
      statusEl.textContent =
        snapshot.components.length === 0
          ? "SWC loaded — no components mounted yet."
          : `${snapshot.components.length} component(s)`;
    }
    renderTree(snapshot.components);
  });
}

refresh();
window.setInterval(refresh, REFRESH_MS);
