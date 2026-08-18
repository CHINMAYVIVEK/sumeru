import { html, type TemplateResult } from "../../template/html.js";
import type { SwcBreadcrumb, SwcViewTab } from "../../types/workspace.js";

export function renderWorkspaceChrome(
  breadcrumbs: SwcBreadcrumb[],
  viewTabs: SwcViewTab[],
): TemplateResult {
  return html`
    <nav class="sum-workspace-chrome" aria-label="Workspace">
      ${breadcrumbs.length > 0
        ? html`<ol class="sum-breadcrumbs">
            ${breadcrumbs.map(
              (b) => html`<li class="sum-breadcrumb">
                ${b.href
                  ? html`<a class="sum-breadcrumb-link" href=${b.href}>${b.label}</a>`
                  : html`<span class="sum-breadcrumb-current">${b.label}</span>`}
              </li>`,
            )}
          </ol>`
        : ""}
      ${viewTabs.length > 0
        ? html`<div class="sum-view-tabs" role="tablist">
            ${viewTabs.map(
              (tab) => html`<a
                class="sum-view-tab${tab.active ? " sum-view-tab--active" : ""}"
                href=${tab.href}
                role="tab"
                aria-selected=${tab.active ? "true" : "false"}
                @click=${(ev: Event) => {
                  ev.preventDefault();
                  window.history.pushState({}, "", tab.href);
                  window.dispatchEvent(new PopStateEvent("popstate"));
                }}
              >
                ${tab.label}
              </a>`,
            )}
          </div>`
        : ""}
    </nav>
  `;
}
