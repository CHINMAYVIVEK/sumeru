/** Template metadata emitted by the sum-template compiler for SWC Vision. */

export interface TemplateSourceMeta {
  /** Component class name */
  component: string;
  /** Source file path (.sum.xml or .ts) */
  file: string;
  /** 1-based line in source */
  line?: number;
  /** Raw template snippet for preview */
  snippet?: string;
}

export function templateMeta(
  component: string,
  file: string,
  opts?: { line?: number; snippet?: string },
): TemplateSourceMeta {
  return { component, file, ...opts };
}
