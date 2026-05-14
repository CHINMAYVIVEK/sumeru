/** In-form notebook: switch visible page per tab. */

export function initNotebookTabs() {
  document.querySelectorAll(".sum-notebook").forEach((root) => {
    const tabs = root.querySelectorAll(".sum-notebook-tab");
    const pages = root.querySelectorAll(".sum-notebook-page");
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
}
