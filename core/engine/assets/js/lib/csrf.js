/** Resolve CSRF token from an element's dataset, page meta tag, or hidden form input. */
export function getCsrfToken(rootEl) {
  if (rootEl?.dataset?.csrf) {
    return rootEl.dataset.csrf;
  }
  return (
    document.querySelector('meta[name="csrf-token"]')?.getAttribute("content") ||
    document.querySelector('input[name="csrf_token"]')?.value ||
    ""
  );
}
