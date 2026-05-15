/**
 * Settings hub (/web/settings): client-side filter for section cards.
 */
(function initSettingsHub() {
  const root = document.querySelector(".sum-settings-hub");
  if (!root) return;

  const input = document.getElementById("sum-settings-hub-q");
  const wrap = document.getElementById("sum-settings-hub-sections");
  if (!input || !wrap) return;

  const sections = () => Array.from(wrap.querySelectorAll(".sum-settings-hub-section"));

  function applyFilter() {
    const q = input.value.trim().toLowerCase();
    const tokens = q ? q.split(/\s+/).filter(Boolean) : [];
    let anyVisible = false;
    for (const sec of sections()) {
      const hay = (sec.getAttribute("data-sum-settings-filter") || "").toLowerCase();
      const match =
        tokens.length === 0 ||
        tokens.every((t) => hay.includes(t));
      sec.classList.toggle("is-filtered-out", !match);
      if (match) anyVisible = true;
    }
    let empty = wrap.querySelector(".sum-settings-hub-filter-empty");
    if (tokens.length > 0 && !anyVisible && sections().length > 0) {
      if (!empty) {
        empty = document.createElement("p");
        empty.className = "sum-settings-hub-filter-empty sum-settings-hub-empty";
        empty.setAttribute("role", "status");
        empty.textContent = "No sections match your search.";
        wrap.appendChild(empty);
      }
    } else if (empty) {
      empty.remove();
    }
  }

  input.addEventListener("input", applyFilter);

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
