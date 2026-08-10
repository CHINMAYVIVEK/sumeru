/**
 * Profile avatar file → resized data-URL into hidden image field + live preview.
 */
const AVATAR_MAX_EDGE = 512;
const AVATAR_MIME = "image/jpeg";
const AVATAR_QUALITY = 0.85;

function showPreview(preview, initials, url) {
  if (!preview || !url) return;
  preview.src = url;
  preview.removeAttribute("hidden");
  preview.hidden = false;
  preview.classList.add("sum-form-avatar-img--visible");
  if (initials) {
    initials.hidden = true;
    initials.setAttribute("hidden", "");
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

export function initAvatarUpload() {
  document.querySelectorAll("[data-sum-avatar]").forEach((root) => {
    const file = root.querySelector("[data-sum-avatar-file]");
    const hidden = root.querySelector("[data-sum-avatar-value]");
    const preview = root.querySelector("[data-sum-avatar-preview]");
    const initials = root.querySelector("[data-sum-avatar-initials]");
    if (!file || !hidden || !preview) return;

    file.addEventListener("change", async () => {
      const f = file.files && file.files[0];
      if (!f || !f.type.startsWith("image/")) return;
      try {
        const url = await resizeToDataURL(f);
        hidden.value = url;
        showPreview(preview, initials, url);
      } catch (err) {
        console.warn("avatar upload failed", err);
      }
    });
  });
}
