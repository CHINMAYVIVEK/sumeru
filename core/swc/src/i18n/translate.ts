const translations = new Map<string, string>();

export function loadTranslations(source: Record<string, string> | undefined): void {
  translations.clear();
  if (!source) return;
  for (const [k, v] of Object.entries(source)) {
    translations.set(k, v);
  }
}

/** Translate a message key (bootstrap translations map). */
export function _t(msg: string): string {
  return translations.get(msg) ?? msg;
}

export function isRTL(): boolean {
  return document.documentElement.dir === "rtl";
}
