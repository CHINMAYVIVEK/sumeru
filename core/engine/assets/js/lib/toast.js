/** Brief toast message on the home dashboard. */
export function showToast(elementId, message, durationMs = 2400) {
  const el = document.getElementById(elementId);
  if (!el) return;
  el.textContent = message;
  el.hidden = false;
  el.classList.add("is-visible");
  clearTimeout(showToast._timers?.[elementId]);
  if (!showToast._timers) showToast._timers = {};
  showToast._timers[elementId] = setTimeout(() => {
    el.classList.remove("is-visible");
    el.hidden = true;
  }, durationMs);
}
