import type { ComponentRecord } from "./bridge.js";
import { getTemplateSource } from "./bridge.js";
import { logRenderEvent } from "./profiler.js";

let panelEl: HTMLElement | null = null;
let selectedId: number | null = null;

export function mountDevtoolsPanel(): void {
  if (typeof window === "undefined") return;
  if (!window.__SWC_DEVTOOLS__) return;
  if (document.getElementById("swc-vision-panel")) return;

  panelEl = document.createElement("aside");
  panelEl.id = "swc-vision-panel";
  panelEl.className = "sum-devtools-panel";
  panelEl.innerHTML = `
    <header class="sum-devtools-header">
      <strong>SWC Vision</strong>
      <button type="button" id="swc-vision-close">×</button>
    </header>
    <section class="sum-devtools-tree" id="swc-vision-tree"></section>
    <section class="sum-devtools-template" id="swc-vision-template"></section>
  `;
  document.body.appendChild(panelEl);

  panelEl.querySelector("#swc-vision-close")?.addEventListener("click", () => {
    panelEl?.remove();
    panelEl = null;
  });

  refreshTree();
  setInterval(refreshTree, 1000);
}

function refreshTree(): void {
  if (!panelEl) return;
  const tree = panelEl.querySelector("#swc-vision-tree");
  const templateView = panelEl.querySelector("#swc-vision-template");
  if (!tree) return;

  const comps = window.__SWC_DEVTOOLS__?.components ?? [];
  tree.innerHTML = comps
    .map(
      (c) =>
        `<button type="button" class="sum-devtools-node${selectedId === c.id ? " sum-devtools-node--active" : ""}" data-id="${c.id}">${c.name} #${c.id}</button>`,
    )
    .join("");

  tree.querySelectorAll("[data-id]").forEach((btn) => {
    btn.addEventListener("click", () => {
      selectedId = Number((btn as HTMLElement).dataset.id);
      showTemplate(comps.find((c) => c.id === selectedId) ?? null, templateView as HTMLElement);
      refreshTree();
    });
  });
}

function showTemplate(comp: ComponentRecord | null, el: HTMLElement | null): void {
  if (!el || !comp) {
    if (el) el.textContent = "Select a component";
    return;
  }
  const meta = getTemplateSource(comp);
  if (!meta) {
    el.innerHTML = `<p>No template metadata for <code>${comp.name}</code></p>`;
    return;
  }
  el.innerHTML = `
    <h4>${meta.component}</h4>
    <p class="sum-devtools-file">${meta.file}${meta.line ? `:${meta.line}` : ""}</p>
    <pre class="sum-devtools-snippet">${meta.snippet ?? ""}</pre>
  `;
}

export function enablePicker(): void {
  document.body.addEventListener(
    "click",
    (ev) => {
      if (!(ev.target instanceof Element)) return;
      if (!ev.altKey) return;
      ev.preventDefault();
      const rec = window.__SWC_DEVTOOLS__?.getComponentForElement(ev.target);
      if (rec) {
        selectedId = rec.id;
        logRenderEvent("pick", rec.name);
        mountDevtoolsPanel();
      }
    },
    true,
  );
}
