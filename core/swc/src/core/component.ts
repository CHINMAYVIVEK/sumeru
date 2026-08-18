import type { ComponentProps } from "./template.js";

export type ComponentConstructor<P extends object = ComponentProps> = new (
  props: P,
  env: import("./env.js").SwcEnv,
) => SwcComponent<P>;

export abstract class SwcComponent<P extends object = ComponentProps> {
  readonly props: P;
  readonly env: import("./env.js").SwcEnv;
  el: HTMLElement | null = null;
  private mounted = false;

  constructor(props: P, env: import("./env.js").SwcEnv) {
    this.props = props;
    this.env = env;
  }

  setup?(): void;

  abstract template(): import("./template.js").TemplateResult;

  onMount?(): void;
  onWillUnmount?(): void;

  render(): HTMLElement {
    const result = this.template();
    const root = result.render();
    this.el = root;
    if (!this.mounted) {
      this.mounted = true;
      this.onMount?.();
    }
    return root;
  }

  patch(): void {
    if (!this.el?.parentElement) return;
    const parent = this.el.parentElement;
    const oldEl = this.el;
    const next = this.template().render();
    parent.replaceChild(next, oldEl);
    this.el = next;
  }

  destroy(): void {
    this.onWillUnmount?.();
    this.el?.remove();
    this.el = null;
    this.mounted = false;
  }
}
