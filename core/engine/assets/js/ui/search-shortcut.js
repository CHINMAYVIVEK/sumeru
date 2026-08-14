/** Press `/` to focus a page search input (when not already typing). */

export function initSearchShortcut(inputId, options = {}) {
  const input = document.getElementById(inputId);
  if (!input) return;

  const skipWhenLauncherOpen = options.skipWhenLauncherOpen !== false;

  document.addEventListener("keydown", (e) => {
    if (e.key !== "/" || e.ctrlKey || e.metaKey || e.altKey) return;
    const tag = document.activeElement && document.activeElement.tagName;
    if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
    if (skipWhenLauncherOpen && document.getElementById("sum-app-launcher")?.open) return;
    e.preventDefault();
    input.focus();
    try {
      input.select();
    } catch (_) {}
  });
}
