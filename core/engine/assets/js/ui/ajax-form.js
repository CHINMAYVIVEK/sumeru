/** Forms marked .sum-ajax-form submit via fetch instead of full navigation. */

export function initAjaxFormCapture() {
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
}
