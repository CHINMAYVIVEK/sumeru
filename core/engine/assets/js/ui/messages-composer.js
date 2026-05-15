/**
 * Activity panel Messages tab: keyboard submit for the chatter composer.
 */
export function initMessagesComposer() {
  const ta = document.getElementById("sum-chatter-body");
  if (!ta) return;
  const form = ta.closest("form.sum-msg-form");
  if (!form) return;

  ta.addEventListener("keydown", (e) => {
    if (e.key !== "Enter") return;
    if (!(e.metaKey || e.ctrlKey)) return;
    if (e.shiftKey) return;
    e.preventDefault();
    if (typeof form.requestSubmit === "function") {
      form.requestSubmit();
    } else {
      form.submit();
    }
  });
}
