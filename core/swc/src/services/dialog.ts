export interface DialogButton {
  label: string;
  primary?: boolean;
  danger?: boolean;
  value?: unknown;
}

export interface DialogOptions {
  title: string;
  body: string;
  buttons?: DialogButton[];
}

export class DialogService {
  private layer: HTMLElement | null = null;

  confirm(title: string, body: string): Promise<boolean> {
    return this.open({
      title,
      body,
      buttons: [
        { label: "Cancel", value: false },
        { label: "OK", primary: true, value: true },
      ],
    }) as Promise<boolean>;
  }

  alert(title: string, body: string): Promise<void> {
    return this.open({
      title,
      body,
      buttons: [{ label: "OK", primary: true, value: true }],
    }).then(() => undefined);
  }

  open(opts: DialogOptions): Promise<unknown> {
    this.close();
    return new Promise((resolve) => {
      const layer = document.createElement("div");
      layer.className = "sum-dialog-layer";
      layer.setAttribute("role", "presentation");

      const dialog = document.createElement("div");
      dialog.className = "sum-dialog";
      dialog.setAttribute("role", "dialog");
      dialog.setAttribute("aria-modal", "true");
      dialog.setAttribute("aria-labelledby", "sum-dialog-title");

      const title = document.createElement("h2");
      title.id = "sum-dialog-title";
      title.className = "sum-dialog-title";
      title.textContent = opts.title;

      const body = document.createElement("p");
      body.className = "sum-dialog-body";
      body.textContent = opts.body;

      const actions = document.createElement("div");
      actions.className = "sum-dialog-actions";

      const buttons = opts.buttons ?? [{ label: "Close", primary: true, value: true }];
      for (const btn of buttons) {
        const el = document.createElement("button");
        el.type = "button";
        el.textContent = btn.label;
        el.className = "sum-dialog-btn";
        if (btn.primary) el.classList.add("sum-dialog-btn--primary");
        if (btn.danger) el.classList.add("sum-dialog-btn--danger");
        el.addEventListener("click", () => {
          this.close();
          resolve(btn.value ?? true);
        });
        actions.appendChild(el);
      }

      dialog.append(title, body, actions);
      layer.appendChild(dialog);
      document.body.appendChild(layer);
      this.layer = layer;

      const onKey = (ev: KeyboardEvent): void => {
        if (ev.key === "Escape") {
          this.close();
          resolve(false);
        }
      };
      document.addEventListener("keydown", onKey, true);
      layer.addEventListener(
        "click",
        (ev) => {
          if (ev.target === layer) {
            this.close();
            resolve(false);
          }
        },
        true,
      );
      layer.addEventListener(
        "remove",
        () => document.removeEventListener("keydown", onKey, true),
        { once: true },
      );

      actions.querySelector("button")?.focus();
    });
  }

  close(): void {
    this.layer?.remove();
    this.layer = null;
  }
}
