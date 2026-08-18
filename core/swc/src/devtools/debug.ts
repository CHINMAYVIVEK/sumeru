/** Debug helpers activated with ?debug=1 in the workspace URL. */

import { initDevtoolsBridge } from "./bridge.js";
import { mountDevtoolsPanel, enablePicker } from "./panel.js";

export function isDebugMode(): boolean {
  return new URLSearchParams(window.location.search).get("debug") === "1";
}

export function mountDebugPanel(): void {
  initDevtoolsBridge();
  if (!isDebugMode()) return;
  if (document.getElementById("sum-debug-panel")) return;
  const el = document.createElement("aside");
  el.id = "sum-debug-panel";
  el.className = "sum-debug-panel";
  el.innerHTML = `<h4>SWC Debug</h4><p>Arch and RPC logging enabled. Alt+click to inspect components.</p>
    <button type="button" id="sum-debug-open-vision">Open SWC Vision</button>`;
  document.body.appendChild(el);
  el.querySelector("#sum-debug-open-vision")?.addEventListener("click", () => mountDevtoolsPanel());
  enablePicker();
}

export function logWorkspacePayload(label: string, payload: unknown): void {
  if (!isDebugMode()) return;
  console.debug(`[SWC ${label}]`, payload);
}

/** Inspect parsed view arch in debug mode. */
export function logViewArch(arch: unknown): void {
  if (!isDebugMode()) return;
  console.debug("[SWC arch]", arch);
}
