/**
 * Profile avatar / image widget file → resized data-URL into hidden field + live preview.
 */
import {
  applyAvatarCropToImg,
  openAvatarCropWizard,
  parseAvatarCrop,
} from "./avatar-crop.js";

const AVATAR_MAX_EDGE = 512;
const AVATAR_MIME = "image/jpeg";
const AVATAR_QUALITY = 0.85;

function showImagePreview(preview, placeholder, url) {
  if (!preview || !url) return;
  preview.src = url;
  preview.removeAttribute("hidden");
  preview.hidden = false;
  if (preview.classList) preview.classList.add("sum-form-avatar-img--visible");
  const thumb = preview.closest?.(".sum-image-thumb");
  if (thumb) {
    thumb.hidden = false;
    thumb.removeAttribute("hidden");
  }
  if (placeholder) {
    placeholder.hidden = true;
    placeholder.setAttribute("hidden", "");
  }
}

function fileToDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ""));
    reader.onerror = () => reject(reader.error || new Error("read failed"));
    reader.readAsDataURL(file);
  });
}

function loadImage(url) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error("image decode failed"));
    img.src = url;
  });
}

async function resizeToDataURL(file) {
  const raw = await fileToDataURL(file);
  try {
    const img = await loadImage(raw);
    const w = img.naturalWidth || img.width;
    const h = img.naturalHeight || img.height;
    if (!w || !h) return raw;
    const scale = Math.min(1, AVATAR_MAX_EDGE / Math.max(w, h));
    const tw = Math.max(1, Math.round(w * scale));
    const th = Math.max(1, Math.round(h * scale));
    const canvas = document.createElement("canvas");
    canvas.width = tw;
    canvas.height = th;
    const ctx = canvas.getContext("2d");
    if (!ctx) return raw;
    ctx.drawImage(img, 0, 0, tw, th);
    return canvas.toDataURL(AVATAR_MIME, AVATAR_QUALITY);
  } catch {
    return raw;
  }
}

function commitAvatar(root, imageUrl, cropJson) {
  const hidden = root.querySelector("[data-sum-avatar-value]");
  const cropHidden = root.querySelector("[data-sum-avatar-crop]");
  const preview = root.querySelector("[data-sum-avatar-preview]");
  const initials = root.querySelector("[data-sum-avatar-initials]");
  if (hidden) hidden.value = imageUrl;
  if (cropHidden) cropHidden.value = cropJson;
  showImagePreview(preview, initials, imageUrl);
  applyAvatarCropToImg(preview, cropJson);
}

function openCropForRoot(root) {
  const hidden = root.querySelector("[data-sum-avatar-value]");
  const cropHidden = root.querySelector("[data-sum-avatar-crop]");
  const url = hidden?.value;
  if (!url) return;
  openAvatarCropWizard({
    imageUrl: url,
    crop: cropHidden?.value || "",
    onConfirm: ({ imageUrl, crop }) => commitAvatar(root, imageUrl, crop),
  });
}

function bindAvatarRoots() {
  document.querySelectorAll("[data-sum-avatar]").forEach((root) => {
    const file = root.querySelector("[data-sum-avatar-file]");
    const hidden = root.querySelector("[data-sum-avatar-value]");
    const cropHidden = root.querySelector("[data-sum-avatar-crop]");
    const preview = root.querySelector("[data-sum-avatar-preview]");
    const initials = root.querySelector("[data-sum-avatar-initials]");
    if (!file || !hidden || !preview) return;

    if (preview.src && cropHidden?.value) {
      applyAvatarCropToImg(preview, cropHidden.value);
    }

    file.addEventListener("change", async () => {
      const f = file.files && file.files[0];
      if (!f || !f.type.startsWith("image/")) return;
      try {
        const url = await resizeToDataURL(f);
        openAvatarCropWizard({
          imageUrl: url,
          crop: "",
          onConfirm: ({ imageUrl, crop }) => commitAvatar(root, imageUrl, crop),
          onCancel: () => {
            file.value = "";
          },
        });
      } catch (err) {
        console.warn("avatar upload failed", err);
      }
    });

    root.querySelector("[data-sum-avatar-adjust]")?.addEventListener("click", () => openCropForRoot(root));
  });
}

function bindImageWidgetRoots() {
  document.querySelectorAll("[data-sum-image]").forEach((root) => {
    const file = root.querySelector("[data-sum-image-file]");
    const hidden = root.querySelector("[data-sum-image-value]");
    const preview = root.querySelector("[data-sum-image-preview]");
    const placeholder = root.querySelector("[data-sum-image-placeholder]");
    if (!file || !hidden) return;

    file.addEventListener("change", async () => {
      const f = file.files && file.files[0];
      if (!f || !f.type.startsWith("image/")) return;
      try {
        const url = await resizeToDataURL(f);
        hidden.value = url;
        showImagePreview(preview, placeholder, url);
      } catch (err) {
        console.warn("image upload failed", err);
      }
    });
  });
}

export function initAvatarUpload() {
  bindAvatarRoots();
  bindImageWidgetRoots();
}
