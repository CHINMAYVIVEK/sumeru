/**
 * Shell: sidebar toggle, activity panel width resize, Messages/Log tabs, notebook tabs.
 * Activity panel: right column, resizable; collapsible with persisted preference.
 */
(function () {
  const shell = document.getElementById("sum-shell");
  if (!shell) return;

  const KEY_SIDEBAR = "sum.shell.sidebarCollapsed";
  const KEY_ACTIVITY_W = "sum.shell.activityWidthPx";
  const KEY_ACTIVITY_HIDDEN = "sum.shell.activityHidden";

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

  function applyActivityWidth(px) {
    document.documentElement.style.setProperty("--sum-activity-width", px + "px");
  }

  function applyActivityHidden(hidden) {
    shell.classList.toggle("sum-shell--activity-hidden", hidden);
    const reveal = document.getElementById("sum-activity-reveal");
    if (reveal) reveal.hidden = !hidden;
  }

  applySidebar(readBool(KEY_SIDEBAR));
  applyActivityWidth(readWidth());
  applyActivityHidden(readBool(KEY_ACTIVITY_HIDDEN));

  function toggleSidebar() {
    const next = !shell.classList.contains("sum-shell--sidebar-collapsed");
    applySidebar(next);
    writeBool(KEY_SIDEBAR, next);
  }

  ["sum-sidebar-toggle", "sum-sidebar-toggle-breadcrumb"].forEach((id) => {
    const el = document.getElementById(id);
    if (el) el.addEventListener("click", toggleSidebar);
  });

  const revealSidebar = document.getElementById("sum-sidebar-reveal");
  if (revealSidebar) {
    revealSidebar.addEventListener("click", () => {
      applySidebar(false);
      writeBool(KEY_SIDEBAR, false);
    });
  }

  function setActivityHidden(hidden) {
    applyActivityHidden(hidden);
    writeBool(KEY_ACTIVITY_HIDDEN, hidden);
  }

  const activityToggle = document.getElementById("sum-activity-toggle");
  if (activityToggle) {
    activityToggle.addEventListener("click", () => {
      setActivityHidden(!shell.classList.contains("sum-shell--activity-hidden"));
    });
  }

  const activityCollapse = document.getElementById("sum-activity-collapse");
  if (activityCollapse) {
    activityCollapse.addEventListener("click", () => setActivityHidden(true));
  }

  const activityReveal = document.getElementById("sum-activity-reveal");
  if (activityReveal) {
    activityReveal.addEventListener("click", () => setActivityHidden(false));
  }

  document.querySelectorAll("[data-sum-activity-tab]").forEach((tab) => {
    tab.addEventListener("click", () => {
      const name = tab.getAttribute("data-sum-activity-tab");
      const panes = { messages: "sum-activity-pane-messages", log: "sum-activity-pane-log" };
      document.querySelectorAll("[data-sum-activity-tab]").forEach((t) => {
        const on = t.getAttribute("data-sum-activity-tab") === name;
        t.classList.toggle("is-active", on);
        t.setAttribute("aria-selected", on ? "true" : "false");
      });
      Object.entries(panes).forEach(([k, id]) => {
        const el = document.getElementById(id);
        if (!el) return;
        el.hidden = k !== name;
      });
    });
  });

  document.querySelectorAll(".sum-notebook").forEach((root) => {
    const tabs = root.querySelectorAll(".sum-notebook-tab");
    const pages = root.querySelectorAll(".o_notebook_page");
    if (tabs.length === 0 || pages.length === 0) return;
    tabs.forEach((tab, i) => {
      tab.addEventListener("click", () => {
        tabs.forEach((t, j) => {
          t.classList.toggle("sum-notebook-tab--active", j === i);
        });
        pages.forEach((p, j) => {
          p.style.display = j === i ? "block" : "none";
        });
      });
    });
  });

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

  // Apps: open read-only module form from grid card or list row (not from action buttons/links).
  if (location.pathname === "/web/apps") {
    document.addEventListener(
      "click",
      (ev) => {
        const card = ev.target.closest(".sum-app-card-prem--open[data-module]");
        const row = ev.target.closest("tr.sum-apps-list-row--open[data-module]");
        const hit = card || row;
        if (!hit) return;
        if (ev.target.closest("form,button,a,.sum-app-card-prem-actions,.sum-apps-list-actions")) return;
        const mod = hit.getAttribute("data-module");
        const layout = hit.getAttribute("data-apps-layout") || "grid";
        if (!mod) return;
        const q = new URLSearchParams();
        q.set("module", mod);
        q.set("layout", layout);
        window.location.href = "/web/apps?" + q.toString();
      },
      true
    );
  }
})();
