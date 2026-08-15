/** Dismissible workspace flash banners (errors, etc.). */
export function initFlashDismiss() {
  document.querySelectorAll(".sum-flash-dismiss").forEach((btn) => {
    btn.addEventListener("click", () => {
      const banner = btn.closest(".sum-flash-prem");
      if (banner) banner.remove();
    });
  });
}
