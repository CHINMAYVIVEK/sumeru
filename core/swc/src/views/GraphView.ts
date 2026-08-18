import { SwcComponent } from "../core/component.js";
import { html } from "../core/template.js";
import type { SwcWorkspacePayload } from "../types/workspace.js";
import { useState, useEffect } from "../core/hooks.js";

interface GraphViewProps {
  payload: SwcWorkspacePayload;
}

/** Minimal v1 graph view over read_group RPC. */
export class GraphView extends SwcComponent<GraphViewProps> {
  private groups: Record<string, unknown>[] = [];
  private measureField = "id";

  setup(): void {
    const [, bump] = useState(0);
    this.bump = () => bump((n) => n + 1);
    useEffect(() => {
      void this.load();
    });
  }

  private bump: (() => void) | null = null;

  private async load(): Promise<void> {
    const p = this.props.payload;
    const groupField = p.arch.fields.find((f) => f.pivotType === "row")?.name ?? "create_date";
    this.measureField = p.arch.fields.find((f) => f.pivotType === "measure")?.name ?? "id";
    this.groups = await this.env.services.rpc.readGroup(p.model, [], [this.measureField], [groupField], 40);
    this.bump?.();
  }

  template() {
    const max = Math.max(...this.groups.map((g) => Number(g[this.measureField] ?? 0)), 1);
    return html`
      <div class="sum-graph-view">
        ${this.groups.map((g) => {
          const label = String(g[`${Object.keys(g).find((k) => k.endsWith("_count")) ?? "name"}`] ?? g.name ?? "");
          const val = Number(g[this.measureField] ?? 0);
          const pct = Math.round((val / max) * 100);
          return html`<div class="sum-graph-bar-row">
            <span class="sum-graph-label">${label}</span>
            <div class="sum-graph-bar" style="width:${pct}%"></div>
            <span class="sum-graph-value">${val}</span>
          </div>`;
        })}
      </div>
    `;
  }
}
