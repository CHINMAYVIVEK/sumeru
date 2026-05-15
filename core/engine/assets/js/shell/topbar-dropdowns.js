/**
 * Ensures only one shell top-bar <details> dropdown is open at a time.
 */
export function initTopbarDropdowns(shell) {
  if (!shell) return;
  const top = shell.querySelector(".sum-topbar-right");
  if (!top) return;
  top.addEventListener("toggle", (ev) => {
    const t = ev.target;
    if (!(t instanceof HTMLDetailsElement) || !t.open) return;
    top.querySelectorAll("details.sum-dropdown").forEach((d) => {
      if (d !== t) d.open = false;
    });
  }, true);
}
