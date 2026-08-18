import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { useState, useEffect } from "../runtime/hooks.js";
import type { SwcBootstrapApp } from "../types/bootstrap.js";

export interface AppLauncherProps {
  apps: SwcBootstrapApp[];
  isOpen: () => boolean;
  requestClose: () => void;
}

export class AppLauncher extends SwcComponent<AppLauncherProps> {
  private query = "";
  private inputBound = false;

  setup(): void {
    const [, bump] = useState(0);
    this.bump = () => bump((n) => n + 1);

    useEffect(() => {
      const dlg = document.getElementById("sum-app-launcher") as HTMLDialogElement | null;
      const onDialogClose = (): void => {
        if (this.props.isOpen()) this.props.requestClose();
      };
      dlg?.addEventListener("close", onDialogClose);

      const onKey = (ev: KeyboardEvent): void => {
        if (ev.key === "Escape" && this.props.isOpen()) {
          ev.preventDefault();
          this.props.requestClose();
        }
      };
      document.addEventListener("keydown", onKey);
      return () => {
        dlg?.removeEventListener("close", onDialogClose);
        document.removeEventListener("keydown", onKey);
      };
    });
  }

  private bump: (() => void) | null = null;

  render(): HTMLElement {
    const el = html`<div hidden data-swc-app-launcher></div>`.render();
    this.el = el;
    this.syncDialog();
    return el;
  }

  private filtered(): SwcBootstrapApp[] {
    const q = this.query.trim().toLowerCase();
    if (!q) return this.props.apps;
    return this.props.apps.filter(
      (a) => a.name.toLowerCase().includes(q) || a.module.toLowerCase().includes(q),
    );
  }

  private syncDialog(): void {
    const dlg = document.getElementById("sum-app-launcher") as HTMLDialogElement | null;
    const input = document.getElementById("sum-app-launcher-input") as HTMLInputElement | null;
    const results = document.getElementById("sum-app-launcher-results") as HTMLUListElement | null;
    if (!dlg || !input || !results) return;

    const open = this.props.isOpen();
    if (!open) {
      if (dlg.open) dlg.close();
      this.query = "";
      input.value = "";
      results.replaceChildren();
      return;
    }

    if (!this.inputBound) {
      this.inputBound = true;
      input.addEventListener("input", () => {
        this.query = input.value;
        this.renderResults(results);
        this.bump?.();
      });
    }

    input.value = this.query;
    this.renderResults(results);

    if (!dlg.open) dlg.showModal();
    queueMicrotask(() => input.focus());
  }

  private renderResults(ul: HTMLUListElement): void {
    ul.replaceChildren();
    for (const app of this.filtered()) {
      const row = document.createElement("li");
      row.className = "sum-app-launcher-result";
      row.setAttribute("role", "option");

      const letter = document.createElement("span");
      letter.className = "sum-app-launcher-result-letter";
      letter.textContent = (app.name.trim()[0] ?? "?").toUpperCase();

      const body = document.createElement("span");
      body.className = "sum-app-launcher-result-body";

      const name = document.createElement("span");
      name.className = "sum-app-launcher-result-name";
      name.textContent = app.name;

      const meta = document.createElement("span");
      meta.className = "sum-app-launcher-result-meta";
      meta.textContent = app.module;

      body.append(name, meta);
      row.append(letter, body);
      row.addEventListener("click", () => {
        this.env.services.action.navigate(app.action);
        this.props.requestClose();
      });
      ul.appendChild(row);
    }
  }

  template() {
    return html`<div hidden data-swc-app-launcher></div>`;
  }
}
