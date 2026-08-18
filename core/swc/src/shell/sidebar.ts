import { KEY_SIDEBAR, readBool, writeBool } from "../util/shell-storage.js";

const CHEVRON_RIGHT =
  '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M9 18l6-6-6-6"/></svg>';

function applySidebar(shell: HTMLElement, collapsed: boolean): void {
  shell.classList.toggle("sum-shell--sidebar-collapsed", collapsed);
  const reveal = document.getElementById("sum-sidebar-reveal");
  if (reveal) reveal.hidden = !collapsed;
}

function paintSidebarRevealIcon(): void {
  const reveal = document.getElementById("sum-sidebar-reveal");
  if (reveal && !reveal.firstChild) reveal.innerHTML = CHEVRON_RIGHT;
}

export function initSidebar(shell: HTMLElement): void {
  applySidebar(shell, readBool(KEY_SIDEBAR));
  paintSidebarRevealIcon();

  const toggleSidebar = (): void => {
    const next = !shell.classList.contains("sum-shell--sidebar-collapsed");
    applySidebar(shell, next);
    writeBool(KEY_SIDEBAR, next);
  };

  for (const id of ["sum-sidebar-toggle", "sum-sidebar-toggle-breadcrumb"]) {
    document.getElementById(id)?.addEventListener("click", toggleSidebar);
  }

  document.getElementById("sum-sidebar-reveal")?.addEventListener("click", () => {
    applySidebar(shell, false);
    writeBool(KEY_SIDEBAR, false);
  });
}
