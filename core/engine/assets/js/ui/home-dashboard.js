/**
 * Home dashboard (/web/home): keyboard shortcut to focus app search.
 */
(function initHomeDashboard() {
  const input = document.getElementById("sum-home-q");
  if (!input) return;

  document.addEventListener("keydown", (e) => {
    if (e.key !== "/" || e.ctrlKey || e.metaKey || e.altKey) return;
    const tag = document.activeElement && document.activeElement.tagName;
    if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
    e.preventDefault();
    input.focus();
    try {
      input.select();
    } catch (_) {}
  });
})();
