import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";

export function renderStubView(title: string, payload: SwcWorkspacePayload) {
  const rows = payload.records ?? [];
  return html`
    <div class="sum-advanced-view">
      <h2>${title}</h2>
      <p class="sum-advanced-view-hint">${rows.length} record(s) loaded.</p>
      <ul>
        ${rows.slice(0, 20).map(
          (row) => html`<li>${String(row.name ?? row.display_name ?? row.id ?? "")}</li>`,
        )}
      </ul>
    </div>
  `;
}
