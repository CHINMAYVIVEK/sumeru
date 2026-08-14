/** Same-origin JSON fetch helpers with optional CSRF header. */
import { getCsrfToken } from "./csrf.js";

export async function getJSON(url) {
  const response = await fetch(url, { credentials: "same-origin" });
  if (!response.ok) {
    throw new Error(`GET ${url} failed: ${response.status}`);
  }
  return response.json();
}

export async function postJSON(url, body, { csrfEl } = {}) {
  const csrfToken = getCsrfToken(csrfEl);
  const response = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrfToken,
    },
    credentials: "same-origin",
    body: JSON.stringify({ ...body, csrf_token: csrfToken }),
  });
  if (!response.ok) {
    throw new Error(`POST ${url} failed: ${response.status}`);
  }
  return response.json();
}
