/**
 * Circular avatar crop wizard for form profile photos.
 */
const DEFAULT_CROP = { x: 50, y: 50, zoom: 1 };
const MIN_ZOOM = 1;
const MAX_ZOOM = 4;

function clamp(n, min, max) {
  return Math.min(max, Math.max(min, n));
}

export function parseAvatarCrop(raw) {
  if (!raw || !String(raw).trim()) return { ...DEFAULT_CROP };
  try {
    const c = JSON.parse(raw);
    return {
      x: clamp(Number(c.x) || 50, 0, 100),
      y: clamp(Number(c.y) || 50, 0, 100),
      zoom: clamp(Number(c.zoom) || 1, MIN_ZOOM, MAX_ZOOM),
    };
  } catch {
    return { ...DEFAULT_CROP };
  }
}

export function serializeAvatarCrop(crop) {
  return JSON.stringify({
    x: Math.round(crop.x * 100) / 100,
    y: Math.round(crop.y * 100) / 100,
    zoom: Math.round(crop.zoom * 1000) / 1000,
  });
}

export function applyAvatarCropToImg(img, crop) {
  if (!img) return;
  const c = parseAvatarCrop(crop);
  img.style.objectPosition = `${c.x}% ${c.y}%`;
  img.style.transform = `scale(${c.zoom})`;
  img.style.transformOrigin = `${c.x}% ${c.y}%`;
  img.classList.add("sum-form-avatar-img--cropped");
}

let modalEl = null;
/** @type {{ imageUrl: string, crop: {x:number,y:number,zoom:number}, onConfirm?: Function, onCancel?: Function, dragging: boolean, lastX: number, lastY: number } | null} */
let state = null;

function ensureModal() {
  if (modalEl) return modalEl;
  modalEl = document.createElement("div");
  modalEl.id = "sum-avatar-crop-modal";
  modalEl.className = "sum-avatar-crop-modal";
  modalEl.hidden = true;
  modalEl.setAttribute("aria-hidden", "true");
  modalEl.innerHTML = `
    <div class="sum-avatar-crop-modal-inner" role="dialog" aria-labelledby="sum-avatar-crop-title">
      <h2 id="sum-avatar-crop-title" class="sum-avatar-crop-title">Adjust profile photo</h2>
      <p class="sum-avatar-crop-hint">Drag to reposition. Use the slider to zoom.</p>
      <div class="sum-avatar-crop-stage" data-sum-crop-stage>
        <div class="sum-avatar-crop-viewport">
          <img class="sum-avatar-crop-img" data-sum-crop-img alt="" draggable="false" />
        </div>
        <div class="sum-avatar-crop-ring" aria-hidden="true"></div>
      </div>
      <label class="sum-avatar-crop-zoom-label">Zoom
        <input type="range" class="sum-avatar-crop-zoom" data-sum-crop-zoom min="1" max="4" step="0.01" value="1" />
      </label>
      <div class="sum-avatar-crop-modal-actions">
        <button type="button" class="sum-avatar-crop-save" data-sum-crop-save>Save</button>
        <button type="button" class="sum-avatar-crop-cancel" data-sum-crop-cancel>Cancel</button>
      </div>
    </div>
  `;
  document.body.appendChild(modalEl);
  bindModalEvents(modalEl);
  return modalEl;
}

function applyCropToStageImg(img, crop) {
  img.style.objectPosition = `${crop.x}% ${crop.y}%`;
  img.style.transform = `scale(${crop.zoom})`;
  img.style.transformOrigin = `${crop.x}% ${crop.y}%`;
}

function hideModal() {
  if (!modalEl) return;
  modalEl.hidden = true;
  modalEl.setAttribute("aria-hidden", "true");
  state = null;
}

function showModal() {
  modalEl.hidden = false;
  modalEl.setAttribute("aria-hidden", "false");
}

function bindModalEvents(modal) {
  if (modal.dataset.bound) return;
  modal.dataset.bound = "1";

  const img = modal.querySelector("[data-sum-crop-img]");
  const zoomInput = modal.querySelector("[data-sum-crop-zoom]");
  const stage = modal.querySelector("[data-sum-crop-stage]");

  zoomInput.addEventListener("input", () => {
    if (!state) return;
    state.crop.zoom = clamp(parseFloat(zoomInput.value) || 1, MIN_ZOOM, MAX_ZOOM);
    applyCropToStageImg(img, state.crop);
  });

  const startDrag = (clientX, clientY) => {
    if (!state) return;
    state.dragging = true;
    state.lastX = clientX;
    state.lastY = clientY;
  };

  const moveDrag = (clientX, clientY) => {
    if (!state || !state.dragging) return;
    const dx = clientX - state.lastX;
    const dy = clientY - state.lastY;
    state.lastX = clientX;
    state.lastY = clientY;
    const factor = 0.15 / state.crop.zoom;
    state.crop.x = clamp(state.crop.x - dx * factor, 0, 100);
    state.crop.y = clamp(state.crop.y - dy * factor, 0, 100);
    applyCropToStageImg(img, state.crop);
  };

  const endDrag = () => {
    if (state) state.dragging = false;
  };

  stage.addEventListener("mousedown", (e) => {
    e.preventDefault();
    startDrag(e.clientX, e.clientY);
  });
  window.addEventListener("mousemove", (e) => moveDrag(e.clientX, e.clientY));
  window.addEventListener("mouseup", endDrag);

  stage.addEventListener(
    "touchstart",
    (e) => {
      const t = e.touches[0];
      if (t) startDrag(t.clientX, t.clientY);
    },
    { passive: true },
  );
  stage.addEventListener(
    "touchmove",
    (e) => {
      const t = e.touches[0];
      if (t) moveDrag(t.clientX, t.clientY);
    },
    { passive: true },
  );
  stage.addEventListener("touchend", endDrag);

  modal.querySelector("[data-sum-crop-save]").addEventListener("click", () => {
    if (!state) return;
    const payload = {
      imageUrl: state.imageUrl,
      crop: serializeAvatarCrop(state.crop),
    };
    const cb = state.onConfirm;
    hideModal();
    cb?.(payload);
  });

  modal.querySelector("[data-sum-crop-cancel]").addEventListener("click", () => {
    const cb = state?.onCancel;
    hideModal();
    cb?.();
  });

  modal.addEventListener("click", (e) => {
    if (e.target === modal) {
      const cb = state?.onCancel;
      hideModal();
      cb?.();
    }
  });
}

/**
 * Open crop wizard. onConfirm receives { imageUrl, crop } where crop is JSON string.
 */
export function openAvatarCropWizard({ imageUrl, crop, onConfirm, onCancel }) {
  const modal = ensureModal();
  const img = modal.querySelector("[data-sum-crop-img]");
  const zoomInput = modal.querySelector("[data-sum-crop-zoom]");

  const initialCrop = parseAvatarCrop(typeof crop === "string" ? crop : serializeAvatarCrop(crop || DEFAULT_CROP));
  state = {
    imageUrl,
    crop: { ...initialCrop },
    onConfirm,
    onCancel,
    dragging: false,
    lastX: 0,
    lastY: 0,
  };

  img.src = imageUrl;
  zoomInput.value = String(state.crop.zoom);
  applyCropToStageImg(img, state.crop);
  showModal();
}
