/** Priority star widget (0–N) on form views. */
export function initPriorityField() {
  document.querySelectorAll(".sum-priority-field").forEach((wrap) => {
    const hidden = wrap.querySelector("[data-sum-priority-value]");
    if (!hidden) return;
    wrap.querySelectorAll(".sum-priority-star[data-priority]").forEach((star) => {
      star.addEventListener("click", () => {
        const val = parseInt(star.dataset.priority || "0", 10);
        hidden.value = String(val);
        wrap.querySelectorAll(".sum-priority-star").forEach((el) => {
          const priority = parseInt(el.dataset.priority || "0", 10);
          el.classList.toggle("sum-priority-star--on", priority <= val);
        });
      });
    });
  });
}
