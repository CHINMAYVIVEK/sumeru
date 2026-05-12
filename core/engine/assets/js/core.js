/**
 * Shell: sidebar toggle, activity panel toggle + width resize, persisted in localStorage.
 */
(function () {
  const shell = document.getElementById("sum-shell");
  if (!shell) return;

  const KEY_SIDEBAR = "sum.shell.sidebarCollapsed";
  const KEY_ACTIVITY = "sum.shell.activityHidden";
  const KEY_ACTIVITY_W = "sum.shell.activityWidthPx";

  function readBool(key) {
    try {
      return localStorage.getItem(key) === "1";
    } catch {
      return false;
    }
  }
  function writeBool(key, v) {
    try {
      localStorage.setItem(key, v ? "1" : "0");
    } catch (_) {}
  }
  function readWidth() {
    try {
      const n = parseInt(localStorage.getItem(KEY_ACTIVITY_W), 10);
      if (n >= 200 && n <= 520) return n;
    } catch (_) {}
    return 300;
  }
  function writeWidth(n) {
    try {
      localStorage.setItem(KEY_ACTIVITY_W, String(Math.round(n)));
    } catch (_) {}
  }

  function applySidebar(collapsed) {
    shell.classList.toggle("sum-shell--sidebar-collapsed", collapsed);
    const reveal = document.getElementById("sum-sidebar-reveal");
    if (reveal) reveal.hidden = !collapsed;
  }

  function applyActivity(hidden) {
    shell.classList.toggle("sum-shell--activity-hidden", hidden);
    const btn = document.getElementById("sum-activity-toggle");
    if (btn) btn.setAttribute("aria-pressed", hidden ? "false" : "true");
  }

  function applyActivityWidth(px) {
    document.documentElement.style.setProperty("--sum-activity-width", px + "px");
  }

  applySidebar(readBool(KEY_SIDEBAR));
  applyActivity(readBool(KEY_ACTIVITY));
  applyActivityWidth(readWidth());

  function toggleSidebar() {
    const next = !shell.classList.contains("sum-shell--sidebar-collapsed");
    applySidebar(next);
    writeBool(KEY_SIDEBAR, next);
  }

  function toggleActivity() {
    const next = !shell.classList.contains("sum-shell--activity-hidden");
    applyActivity(next);
    writeBool(KEY_ACTIVITY, next);
  }

  ["sum-sidebar-toggle", "sum-sidebar-toggle-breadcrumb"].forEach((id) => {
    const el = document.getElementById(id);
    if (el) el.addEventListener("click", toggleSidebar);
  });

  const reveal = document.getElementById("sum-sidebar-reveal");
  if (reveal) {
    reveal.addEventListener("click", () => {
      applySidebar(false);
      writeBool(KEY_SIDEBAR, false);
    });
  }

  const actBtn = document.getElementById("sum-activity-toggle");
  if (actBtn) actBtn.addEventListener("click", toggleActivity);

  const resizer = document.getElementById("sum-activity-resizer");
  if (resizer) {
    let dragging = false;
    let startX = 0;
    let startW = 300;

    resizer.addEventListener("mousedown", (e) => {
      if (shell.classList.contains("sum-shell--activity-hidden")) return;
      dragging = true;
      startX = e.clientX;
      startW = readWidth();
      document.body.style.cursor = "col-resize";
      document.body.style.userSelect = "none";
      e.preventDefault();
    });

    window.addEventListener("mousemove", (e) => {
      if (!dragging) return;
      const delta = startX - e.clientX;
      let w = startW + delta;
      w = Math.min(520, Math.max(200, w));
      applyActivityWidth(w);
    });

    window.addEventListener("mouseup", () => {
      if (!dragging) return;
      dragging = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      const m = document.documentElement.style.getPropertyValue("--sum-activity-width");
      const px = parseInt(m, 10);
      if (!Number.isNaN(px)) writeWidth(px);
    });
  }

  document.addEventListener("submit", async (e) => {
    const form = e.target;
    if (!(form instanceof HTMLFormElement)) return;
    if (!form.classList.contains("sum-ajax-form")) return;

    e.preventDefault();
    const data = new FormData(form);
    try {
      await fetch(form.action || "/save", {
        method: "POST",
        body: data,
      });
      alert("Saved!");
    } catch (err) {
      console.error(err);
      alert("Request failed.");
    }
  });
})();
