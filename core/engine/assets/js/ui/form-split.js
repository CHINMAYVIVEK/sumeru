/**
 * Resizable two-pane form sheet: left summary column vs main fields (localStorage width).
 */
const KEY = "sum.form.sheetChromeWidthPx";

function clamp(n, lo, hi) {
  return Math.min(hi, Math.max(lo, n));
}

export function initFormSplit() {
  const root = document.querySelector("[data-sum-form-split]");
  if (!root) return;

  const left = root.querySelector(".sum-form-split-left");
  const resizer = root.querySelector(".sum-form-split-resizer");
  if (!left || !resizer) return;

  function readWidth() {
    try {
      const n = parseInt(localStorage.getItem(KEY), 10);
      if (n >= 200 && n <= 480) return n;
    } catch (_) {}
    return 280;
  }

  function apply(px) {
    const w = clamp(px, 200, 480);
    root.style.setProperty("--sum-form-chrome-width", `${w}px`);
  }

  apply(readWidth());

  let dragging = false;
  let startX = 0;
  let startW = readWidth();

  resizer.addEventListener("mousedown", (e) => {
    dragging = true;
    startX = e.clientX;
    startW = readWidth();
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
    e.preventDefault();
  });

  window.addEventListener("mousemove", (e) => {
    if (!dragging) return;
    const delta = e.clientX - startX;
    apply(startW + delta);
  });

  window.addEventListener("mouseup", () => {
    if (!dragging) return;
    dragging = false;
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
    try {
      const v = root.style.getPropertyValue("--sum-form-chrome-width");
      const px = parseInt(v, 10);
      if (!Number.isNaN(px)) localStorage.setItem(KEY, String(px));
    } catch (_) {}
  });

  resizer.addEventListener(
    "keydown",
    (e) => {
      const step = e.shiftKey ? 20 : 8;
      let w = readWidth();
      if (e.key === "ArrowLeft") w -= step;
      else if (e.key === "ArrowRight") w += step;
      else return;
      e.preventDefault();
      apply(w);
      try {
        localStorage.setItem(KEY, String(clamp(w, 200, 480)));
      } catch (_) {}
    },
    true
  );
}
