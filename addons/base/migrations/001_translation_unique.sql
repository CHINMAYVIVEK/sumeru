-- Unique key for i18n import upserts (lang + source + module).
CREATE UNIQUE INDEX IF NOT EXISTS sys_translation_lang_src_module_uidx
  ON sys_translation (lang, src, module);
