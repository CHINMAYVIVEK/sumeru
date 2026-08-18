"use strict";
var SumeruSWC = (() => {
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __export = (target, all) => {
    for (var name in all)
      __defProp(target, name, { get: all[name], enumerable: true });
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

  // src/main.ts
  var main_exports = {};
  __export(main_exports, {
    SwcApp: () => SwcApp,
    SwcEnv: () => SwcEnv,
    registry: () => registry
  });

  // src/core/component.ts
  var SwcComponent = class {
    props;
    env;
    el = null;
    mounted = false;
    constructor(props, env) {
      this.props = props;
      this.env = env;
    }
    render() {
      const result = this.template();
      const root = result.render();
      this.el = root;
      if (!this.mounted) {
        this.mounted = true;
        this.onMount?.();
      }
      return root;
    }
    patch() {
      if (!this.el?.parentElement) return;
      const parent = this.el.parentElement;
      const oldEl = this.el;
      const next = this.template().render();
      parent.replaceChild(next, oldEl);
      this.el = next;
    }
    destroy() {
      this.onWillUnmount?.();
      this.el?.remove();
      this.el = null;
      this.mounted = false;
    }
  };

  // src/core/hooks.ts
  var mountCallbacks = [];
  var unmountCallbacks = [];
  var activeComponent = null;
  function setActiveComponent(comp) {
    activeComponent = comp;
  }
  function runMountCallbacks() {
    for (const fn of mountCallbacks.splice(0)) {
      fn();
    }
  }
  function runUnmountCallbacks() {
    for (const fn of unmountCallbacks.splice(0)) {
      fn();
    }
  }
  function onMount(fn) {
    mountCallbacks.push(fn);
  }
  function onWillUnmount(fn) {
    unmountCallbacks.push(fn);
  }
  function useState(initial) {
    let value = initial;
    const set = (next) => {
      value = typeof next === "function" ? next(value) : next;
      activeComponent?.schedulePatch();
    };
    return [() => value, set];
  }
  function useEffect(fn) {
    onMount(() => {
      const cleanup = fn();
      if (typeof cleanup === "function") {
        onWillUnmount(cleanup);
      }
    });
  }

  // src/core/template.ts
  var VOID_ELEMENTS = /* @__PURE__ */ new Set([
    "area",
    "base",
    "br",
    "col",
    "embed",
    "hr",
    "img",
    "input",
    "link",
    "meta",
    "param",
    "source",
    "track",
    "wbr"
  ]);
  function escapeHtml(text) {
    return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
  }
  function flattenValues(values) {
    const out = [];
    for (const v of values) {
      if (v == null || v === false) continue;
      if (typeof v === "object" && "render" in v && typeof v.render === "function") {
        out.push(v.render());
        continue;
      }
      if (v instanceof HTMLElement) {
        out.push(v);
        continue;
      }
      if (Array.isArray(v)) {
        out.push(...flattenValues(v));
        continue;
      }
      out.push(String(v));
    }
    return out;
  }
  function awaitingAttrName(partial) {
    const m = partial.match(/([@\.\w-]+)=\s*$/);
    return m ? m[1] : null;
  }
  function stripAwaitingAttr(partial) {
    const attr = awaitingAttrName(partial);
    if (!attr) return partial;
    return partial.slice(0, partial.length - attr.length - 1);
  }
  function isAttrWhitespace(ch) {
    return ch === " " || ch === "\n" || ch === "\r" || ch === "	";
  }
  function parseTagAttributes(raw) {
    const s = raw.trim();
    const tagMatch = s.match(/^([^\s/]+)/);
    const tag = tagMatch?.[1] || "div";
    const rest = s.slice(tag.length).trim();
    const attrs = {};
    let i = 0;
    while (i < rest.length) {
      while (i < rest.length && isAttrWhitespace(rest[i])) i++;
      if (i >= rest.length) break;
      const keyStart = i;
      while (i < rest.length && rest[i] !== "=" && !isAttrWhitespace(rest[i])) i++;
      const key = rest.slice(keyStart, i).trim();
      if (!key) break;
      if (i >= rest.length || rest[i] !== "=") {
        attrs[key] = "";
        continue;
      }
      i++;
      if (i >= rest.length) {
        attrs[key] = "";
        break;
      }
      if (rest[i] === '"' || rest[i] === "'") {
        const quote = rest[i++];
        let val2 = "";
        while (i < rest.length && rest[i] !== quote) val2 += rest[i++];
        if (i < rest.length) i++;
        attrs[key] = val2;
        continue;
      }
      let val = "";
      while (i < rest.length && !isAttrWhitespace(rest[i])) val += rest[i++];
      attrs[key] = val;
    }
    return { tag, attrs };
  }
  function buildTree(strings, values) {
    const root = [];
    const stack = [];
    let partial = null;
    let partialHandlers = {};
    const parentChildren = () => stack.length > 0 ? stack[stack.length - 1].children : root;
    const appendChild = (child) => {
      parentChildren().push(child);
    };
    const openTag = (raw, handlers = {}) => {
      const trimmed = raw.trim();
      const selfClosing = trimmed.endsWith("/");
      const inner = selfClosing ? trimmed.slice(0, -1).trim() : trimmed;
      const parsed = parseTagAttributes(inner);
      const node = {
        tag: parsed.tag,
        attrs: parsed.attrs,
        handlers: { ...handlers },
        children: []
      };
      appendChild(node);
      if (!selfClosing && !VOID_ELEMENTS.has(parsed.tag.toLowerCase())) {
        stack.push(node);
      }
    };
    const closeTag = (raw) => {
      const name = raw.slice(1).trim().split(/\s+/)[0]?.toLowerCase() ?? "";
      if (!name) return;
      while (stack.length > 0) {
        const top = stack[stack.length - 1];
        if (top.tag.toLowerCase() === name) {
          stack.pop();
          break;
        }
        stack.pop();
      }
    };
    const completePartial = (rest) => {
      if (partial === null) return;
      let raw = partial + rest;
      if (raw.startsWith("<")) raw = raw.slice(1);
      openTag(raw, partialHandlers);
      partial = null;
      partialHandlers = {};
    };
    const processText = (text) => {
      let i = 0;
      while (i < text.length) {
        if (partial !== null) {
          const gt2 = text.indexOf(">", i);
          if (gt2 === -1) {
            partial += text.slice(i);
            return;
          }
          completePartial(text.slice(i, gt2));
          i = gt2 + 1;
          continue;
        }
        const lt = text.indexOf("<", i);
        if (lt === -1) {
          const tail = text.slice(i);
          if (tail) appendChild(tail);
          return;
        }
        if (lt > i) {
          appendChild(text.slice(i, lt));
        }
        const gt = text.indexOf(">", lt);
        if (gt === -1) {
          partial = text.slice(lt);
          return;
        }
        const inner = text.slice(lt + 1, gt);
        i = gt + 1;
        if (inner.startsWith("/")) {
          closeTag(inner);
        } else {
          openTag(inner);
        }
      }
    };
    const processValue = (value) => {
      if (partial !== null) {
        const attr = awaitingAttrName(partial);
        if (value == null || value === false) {
          if (attr) partial = stripAwaitingAttr(partial);
          return;
        }
        if (typeof value === "function") {
          if (attr) partialHandlers[attr.startsWith("@") ? attr.slice(1) : attr] = value;
          return;
        }
        const text = String(value);
        const needsQuotes = attr && !attr.startsWith("@") && (text === "" || text.includes(" ") || text.includes("\n") || text.includes("	") || text.includes(".") || text.includes("=") || text.includes("<") || text.includes(">"));
        if (needsQuotes) {
          partial += `"${text.replace(/"/g, "&quot;")}"`;
        } else {
          partial += text;
        }
        return;
      }
      for (const child of flattenValues([value])) {
        appendChild(child);
      }
    };
    for (let i = 0; i < strings.length; i++) {
      if (strings[i]) processText(strings[i]);
      if (i < values.length) processValue(values[i]);
    }
    if (partial) appendChild(partial);
    return root;
  }
  function html(strings, ...values) {
    const vnodes = buildTree(strings, values);
    return {
      render() {
        const root = document.createElement("div");
        root.style.display = "contents";
        for (const node of vnodes) {
          if (typeof node === "string") {
            root.appendChild(document.createTextNode(node));
            continue;
          }
          if (node instanceof HTMLElement) {
            root.appendChild(node);
            continue;
          }
          root.appendChild(renderVNode(node));
        }
        if (root.childNodes.length === 1 && root.firstElementChild) {
          return root.firstElementChild;
        }
        return root;
      }
    };
  }
  function applyStyle(el, raw) {
    for (const part of raw.split(";")) {
      const idx = part.indexOf(":");
      if (idx === -1) continue;
      const prop = part.slice(0, idx).trim();
      const val = part.slice(idx + 1).trim();
      if (prop) el.style.setProperty(prop, val);
    }
  }
  function renderVNode(vn) {
    const el = document.createElement(vn.tag);
    for (const [k, v] of Object.entries(vn.attrs)) {
      if (k.startsWith("@")) {
        continue;
      }
      if (k.startsWith(".")) {
        el.classList.add(k.slice(1));
      } else if (k === "class" && v) {
        for (const c of v.split(/\s+/)) {
          if (c) el.classList.add(c);
        }
      } else if (k === "style" && v) {
        applyStyle(el, v);
      } else if (k.startsWith("data-") || k === "id" || k === "for" || k === "href" || k === "type" || k === "name" || k === "value" || k === "placeholder" || k === "aria-label" || k === "title" || k === "role" || k === "aria-selected" || k === "checked" || k === "src" || k === "alt" || k === "rows" || k === "selected" || k === "method" || k === "action" || k === "enctype" || k === "accept" || k === "open" || k === "hidden" || k === "disabled") {
        el.setAttribute(k, v);
      }
    }
    for (const [event, handler] of Object.entries(vn.handlers)) {
      el.addEventListener(event, handler);
    }
    for (const child of vn.children) {
      if (typeof child === "string") {
        el.insertAdjacentHTML("beforeend", escapeHtml(child));
      } else if (child instanceof HTMLElement) {
        el.appendChild(child);
      } else {
        el.appendChild(renderVNode(child));
      }
    }
    return el;
  }

  // src/core/error.ts
  var SwcError = class extends Error {
    code;
    details;
    constructor(message, code = "swc_error", details) {
      super(message);
      this.name = "SwcError";
      this.code = code;
      this.details = details;
    }
  };

  // src/core/app.ts
  var ErrorBoundary = class extends SwcComponent {
    template() {
      const { error, retry } = this.props;
      return html`
      <div class="sum-flash sum-flash--error">
        <strong>SWC error</strong>
        <p>${error.message}</p>
        <button type="button" class="sum-btn sum-btn--secondary" @click=${() => retry()}>Retry</button>
      </div>
    `;
    }
  };
  function renderErrorFallback(error, retry) {
    const wrap = document.createElement("div");
    wrap.className = "sum-flash sum-flash--error";
    const title = document.createElement("strong");
    title.textContent = "SWC error";
    wrap.appendChild(title);
    const message = document.createElement("p");
    message.textContent = error.message;
    wrap.appendChild(message);
    const button = document.createElement("button");
    button.type = "button";
    button.className = "sum-btn sum-btn--secondary";
    button.textContent = "Retry";
    button.addEventListener("click", retry);
    wrap.appendChild(button);
    return wrap;
  }
  function showError(rootEl, env, error, retry) {
    try {
      const boundary = new ErrorBoundary({ error, retry }, env);
      rootEl.replaceChildren(boundary.render());
    } catch {
      rootEl.replaceChildren(renderErrorFallback(error, retry));
    }
  }
  var SwcApp = class _SwcApp {
    env;
    Root;
    rootEl = null;
    component = null;
    scheduled = false;
    constructor(env, Root) {
      this.env = env;
      this.Root = Root;
    }
    static start(mountEl, env, Root) {
      const app = new _SwcApp(env, Root);
      app.mount(mountEl);
      return app;
    }
    mount(el) {
      this.rootEl = el;
      this.renderRoot();
    }
    schedulePatch() {
      if (this.scheduled) return;
      this.scheduled = true;
      requestAnimationFrame(() => {
        this.scheduled = false;
        this.renderRoot();
      });
    }
    renderRoot() {
      if (!this.rootEl) return;
      try {
        if (!this.component) {
          this.component = new this.Root({}, this.env);
          this.component.setup?.();
          setActiveComponent({ schedulePatch: () => this.schedulePatch() });
          runMountCallbacks();
          this.rootEl.replaceChildren(this.component.render());
        } else {
          setActiveComponent({ schedulePatch: () => this.schedulePatch() });
          this.component.patch();
        }
      } catch (err) {
        const swcErr = err instanceof SwcError ? err : new SwcError(String(err));
        showError(this.rootEl, this.env, swcErr, () => this.retry());
      }
    }
    retry() {
      runUnmountCallbacks();
      this.component?.destroy();
      this.component = null;
      this.renderRoot();
    }
    destroy() {
      runUnmountCallbacks();
      this.component?.destroy();
      this.component = null;
      this.rootEl = null;
    }
  };

  // src/core/env.ts
  var SwcEnv = class {
    bootstrap;
    services;
    constructor(bootstrap2, services) {
      this.bootstrap = bootstrap2;
      this.services = services;
    }
    get(name) {
      return this.services[name];
    }
  };

  // src/types/bootstrap.ts
  function readBootstrap() {
    const boot = window.__SWC_BOOTSTRAP__;
    if (!boot) {
      throw new Error("SWC bootstrap missing on window.__SWC_BOOTSTRAP__");
    }
    return boot;
  }

  // src/services/rpc.ts
  var RpcService = class {
    url;
    csrfToken;
    searchReadCache = /* @__PURE__ */ new Map();
    constructor(url, csrfToken) {
      this.url = url;
      this.csrfToken = csrfToken;
    }
    searchReadKey(model, domain, fields, limit) {
      return JSON.stringify({ model, domain, fields, limit });
    }
    /** Clears cached search_read results (e.g. after writes). */
    invalidateSearchReadCache() {
      this.searchReadCache.clear();
    }
    async dispatch(model, method, args = [], kwargs = {}) {
      const body = { model, method, args };
      if (Object.keys(kwargs).length > 0) {
        body.kwargs = kwargs;
      }
      const res = await fetch(this.url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": this.csrfToken
        },
        credentials: "same-origin",
        body: JSON.stringify(body)
      });
      if (!res.ok) {
        throw new SwcError(`RPC HTTP ${res.status}`, "rpc_http");
      }
      const data = await res.json();
      if (data.ok === false || data.error) {
        throw new SwcError(data.error?.message ?? "RPC failed", "rpc_error", data.error);
      }
      return data.result;
    }
    searchRead(model, domain = [], fields = [], limit = 80) {
      const key = this.searchReadKey(model, domain, fields, limit);
      let pending = this.searchReadCache.get(key);
      if (!pending) {
        pending = this.dispatch(model, "search_read", [domain, fields], {
          limit
        });
        this.searchReadCache.set(key, pending);
        void pending.catch(() => {
          this.searchReadCache.delete(key);
        });
      }
      return pending;
    }
    read(model, ids, fields = []) {
      return this.dispatch(model, "read", [ids, fields]);
    }
    write(model, ids, values) {
      this.invalidateSearchReadCache();
      return this.dispatch(model, "write", [ids, values]);
    }
    create(model, values) {
      this.invalidateSearchReadCache();
      return this.dispatch(model, "create", [values]);
    }
    unlink(model, ids) {
      this.invalidateSearchReadCache();
      return this.dispatch(model, "unlink", [ids]);
    }
    callMethod(model, method, recordId) {
      return this.dispatch(model, "call", [recordId, method]);
    }
    readGroup(model, domain, fields, groupBy, limit = 80) {
      const spec = {
        domain,
        groupby: groupBy,
        fields: fields.map((name) => ({ name, field: name, measure: name === "id" ? "count" : "sum" }))
      };
      return this.dispatch(model, "read_group", [spec], { limit });
    }
  };

  // src/services/http.ts
  var HttpService = class {
    csrfToken;
    constructor(csrfToken) {
      this.csrfToken = csrfToken;
    }
    get csrf() {
      return this.csrfToken;
    }
    async getJSON(url) {
      const res = await fetch(url, {
        credentials: "same-origin",
        headers: { Accept: "application/json" }
      });
      if (!res.ok) {
        throw new SwcError(`GET ${url} failed: ${res.status}`, "http_get");
      }
      return await res.json();
    }
    async postForm(url, data) {
      const body = new URLSearchParams({ ...data, csrf_token: this.csrfToken });
      return fetch(url, {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body
      });
    }
    async postJSON(url, body) {
      const payload = { ...body, csrf_token: this.csrfToken };
      const res = await fetch(url, {
        method: "POST",
        credentials: "same-origin",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
          "X-CSRF-Token": this.csrfToken
        },
        body: JSON.stringify(payload)
      });
      if (!res.ok) {
        throw new SwcError(`POST ${url} failed: ${res.status}`, "http_post");
      }
      return await res.json();
    }
  };

  // src/services/notification.ts
  var NotificationService = class {
    stack;
    constructor(stackEl) {
      this.stack = stackEl ?? document.getElementById("sum-toast-stack") ?? this.createStack();
    }
    createStack() {
      const el = document.createElement("div");
      el.id = "sum-toast-stack";
      el.className = "sum-toast-stack";
      el.setAttribute("aria-live", "polite");
      document.body.appendChild(el);
      return el;
    }
    show(msg, timeoutMs = 6e3) {
      const toast = document.createElement("div");
      toast.className = `sum-toast sum-toast--${msg.kind || "info"}`;
      toast.innerHTML = `<strong>${escape(msg.title)}</strong><span>${escape(msg.body)}</span>`;
      this.stack.appendChild(toast);
      window.setTimeout(() => toast.remove(), timeoutMs);
    }
    bootstrap(messages) {
      for (const m of messages ?? []) {
        this.show(m);
      }
    }
  };
  function escape(s) {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }

  // src/services/action.ts
  var ActionService = class {
    navigate(url) {
      window.location.assign(url);
    }
    openWindowAction(actionId, menuId, extra) {
      const params = new URLSearchParams({ action: String(actionId) });
      if (menuId) params.set("menu_id", menuId);
      for (const [k, v] of Object.entries(extra ?? {})) {
        if (v) params.set(k, v);
      }
      this.navigate(`/web?${params.toString()}`);
    }
    openRecord(_model, actionId, menuId, recordId, viewType = "form") {
      const params = new URLSearchParams({
        action: String(actionId),
        menu_id: menuId,
        view_type: viewType,
        id: String(recordId)
      });
      this.navigate(`/web?${params.toString()}`);
    }
  };

  // src/services/router.ts
  var RouterService = class {
    parse(location = window.location) {
      const q = new URLSearchParams(location.search);
      return {
        actionId: Number(q.get("action") ?? "0"),
        menuId: q.get("menu_id") ?? "",
        viewType: q.get("view_type") ?? "",
        recordId: Number(q.get("id") ?? "0"),
        formEdit: q.get("edit") === "1",
        listSearch: q.get("q") ?? ""
      };
    }
    workspaceUrl(route) {
      const current = this.parse();
      const merged = { ...current, ...route };
      const params = new URLSearchParams();
      if (merged.actionId) params.set("action", String(merged.actionId));
      if (merged.menuId) params.set("menu_id", merged.menuId);
      if (merged.viewType) params.set("view_type", merged.viewType);
      if (merged.recordId) params.set("id", String(merged.recordId));
      if (merged.formEdit) params.set("edit", "1");
      if (merged.listSearch) params.set("q", merged.listSearch);
      return `/web?${params.toString()}`;
    }
    push(route) {
      const url = this.workspaceUrl(route);
      window.history.pushState({}, "", url);
      window.dispatchEvent(new PopStateEvent("popstate"));
    }
  };

  // src/services/bus.ts
  var BusService = class {
    handlers = /* @__PURE__ */ new Map();
    ws = null;
    subscribe(channel, handler) {
      if (!this.handlers.has(channel)) {
        this.handlers.set(channel, /* @__PURE__ */ new Set());
      }
      this.handlers.get(channel).add(handler);
      return () => this.handlers.get(channel)?.delete(handler);
    }
    emit(channel, payload) {
      for (const fn of this.handlers.get(channel) ?? []) {
        fn(payload);
      }
    }
    /** Connect to /web/swc/bus when bootstrap.busEnabled is true. */
    connect(url = "/web/swc/bus") {
      if (this.ws) return;
      try {
        const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
        this.ws = new WebSocket(`${proto}//${window.location.host}${url}`);
        this.ws.addEventListener("message", (ev) => {
          try {
            const msg = JSON.parse(String(ev.data));
            if (msg.channel) this.emit(msg.channel, msg.payload);
          } catch {
          }
        });
        this.ws.addEventListener("close", () => {
          this.ws = null;
        });
      } catch {
      }
    }
    disconnect() {
      this.ws?.close();
      this.ws = null;
    }
  };

  // src/shell/AppLauncher.ts
  var AppLauncher = class extends SwcComponent {
    query = "";
    inputBound = false;
    setup() {
      const [, bump] = useState(0);
      this.bump = () => bump((n) => n + 1);
      useEffect(() => {
        const dlg = document.getElementById("sum-app-launcher");
        const onDialogClose = () => {
          if (this.props.isOpen()) this.props.requestClose();
        };
        dlg?.addEventListener("close", onDialogClose);
        const onKey = (ev) => {
          if (ev.key === "Escape" && this.props.isOpen()) {
            ev.preventDefault();
            this.props.requestClose();
          }
        };
        document.addEventListener("keydown", onKey);
        return () => {
          dlg?.removeEventListener("close", onDialogClose);
          document.removeEventListener("keydown", onKey);
        };
      });
    }
    bump = null;
    render() {
      const el = html`<div hidden data-swc-app-launcher></div>`.render();
      this.el = el;
      this.syncDialog();
      return el;
    }
    filtered() {
      const q = this.query.trim().toLowerCase();
      if (!q) return this.props.apps;
      return this.props.apps.filter(
        (a) => a.name.toLowerCase().includes(q) || a.module.toLowerCase().includes(q)
      );
    }
    syncDialog() {
      const dlg = document.getElementById("sum-app-launcher");
      const input = document.getElementById("sum-app-launcher-input");
      const results = document.getElementById("sum-app-launcher-results");
      if (!dlg || !input || !results) return;
      const open = this.props.isOpen();
      if (!open) {
        if (dlg.open) dlg.close();
        this.query = "";
        input.value = "";
        results.replaceChildren();
        return;
      }
      if (!this.inputBound) {
        this.inputBound = true;
        input.addEventListener("input", () => {
          this.query = input.value;
          this.renderResults(results);
          this.bump?.();
        });
      }
      input.value = this.query;
      this.renderResults(results);
      if (!dlg.open) dlg.showModal();
      queueMicrotask(() => input.focus());
    }
    renderResults(ul) {
      ul.replaceChildren();
      for (const app of this.filtered()) {
        const row = document.createElement("li");
        row.className = "sum-app-launcher-result";
        row.setAttribute("role", "option");
        const letter = document.createElement("span");
        letter.className = "sum-app-launcher-result-letter";
        letter.textContent = (app.name.trim()[0] ?? "?").toUpperCase();
        const body = document.createElement("span");
        body.className = "sum-app-launcher-result-body";
        const name = document.createElement("span");
        name.className = "sum-app-launcher-result-name";
        name.textContent = app.name;
        const meta = document.createElement("span");
        meta.className = "sum-app-launcher-result-meta";
        meta.textContent = app.module;
        body.append(name, meta);
        row.append(letter, body);
        row.addEventListener("click", () => {
          this.env.services.action.navigate(app.action);
          this.props.requestClose();
        });
        ul.appendChild(row);
      }
    }
    template() {
      return html`<div hidden data-swc-app-launcher></div>`;
    }
  };

  // src/views/view-toolbar.ts
  function linkButton(href, label, className = "sum-btn sum-btn--secondary") {
    const a = document.createElement("a");
    a.className = className;
    a.href = href;
    a.textContent = label;
    return a;
  }
  function visibleFieldNames(fields) {
    return fields.map((f) => f.name).filter(Boolean).join(",");
  }
  function newRecordUrl(payload) {
    const params = new URLSearchParams();
    if (payload.actionId > 0) params.set("action", String(payload.actionId));
    if (payload.menuId) params.set("menu_id", payload.menuId);
    params.set("view_type", "form");
    return `/web?${params.toString()}`;
  }
  function exportQuery(payload, fields, recordId = 0) {
    const params = new URLSearchParams();
    params.set("model", payload.model);
    if (payload.actionId > 0) params.set("action", String(payload.actionId));
    if (fields) params.set("fields", fields);
    if (recordId > 0) params.set("id", String(recordId));
    return params;
  }
  function renderNewButton(payload) {
    return linkButton(newRecordUrl(payload), "New", "sum-btn sum-list-btn-new");
  }
  function toolbarButton(label, className, onClick, disabled = false) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = className;
    btn.textContent = label;
    btn.disabled = disabled;
    btn.addEventListener("click", onClick);
    return btn;
  }
  function resolveHeaderButtonClass(archClass) {
    const base = "sum-header-btn";
    if (archClass?.includes("sum_highlight")) {
      return `${base} sum-header-btn--primary`;
    }
    return `${base} sum-header-btn--secondary`;
  }
  function headerButton(label, archClass, onClick, disabled = false) {
    const className = disabled ? `${resolveHeaderButtonClass(archClass)} sum-header-btn--disabled` : resolveHeaderButtonClass(archClass);
    return toolbarButton(label, className, onClick, disabled);
  }
  function renderReportActions(payload, fields, recordId = 0) {
    const report = payload.arch.report;
    if (!report?.download && !report?.upload) return null;
    const exportParams = exportQuery(payload, fields, recordId);
    const items = [];
    if (report.download) {
      items.push(linkButton(`/web/export/csv?${exportParams.toString()}`, "Export CSV"));
      items.push(linkButton(`/web/export/pdf?${exportParams.toString()}`, "Export PDF"));
    }
    if (report.upload && fields) {
      const templateParams = new URLSearchParams(exportParams);
      items.push(linkButton(`/web/bulk/template?${templateParams.toString()}`, "Download template"));
      items.push(
        html`<form class="sum-list-upload-form" method="post" enctype="multipart/form-data" action="/web/bulk/upload">
        <input type="hidden" name="csrf_token" value=${payload.csrfToken} />
        <input type="hidden" name="model" value=${payload.model} />
        ${payload.actionId > 0 ? html`<input type="hidden" name="action" value=${String(payload.actionId)} />` : ""}
        <input type="hidden" name="fields" value=${fields} />
        <label class="sum-btn sum-btn--secondary sum-list-upload-label">
          Import CSV
          <input type="file" name="file" accept=".csv,text/csv" class="sum-list-upload-input" @change=${(ev) => ev.target.form?.requestSubmit()} />
        </label>
      </form>`
      );
    }
    if (items.length === 0) return null;
    const wrap = document.createElement("div");
    wrap.className = "sum-view-toolbar-actions";
    for (const item of items) {
      wrap.appendChild(item instanceof HTMLElement ? item : item.render());
    }
    return wrap;
  }

  // src/views/ListView.ts
  var ListView = class extends SwcComponent {
    search = "";
    setup() {
      this.search = this.props.payload.listSearch ?? "";
      const [, bump] = useState(0);
      this.bump = () => bump((n) => n + 1);
    }
    bump = null;
    columns() {
      return this.props.payload.arch.fields.filter((f) => !f.invisible);
    }
    rows() {
      return this.props.payload.records ?? [];
    }
    applySearch() {
      const url = this.env.services.router.workspaceUrl({
        actionId: this.props.payload.actionId,
        menuId: this.props.payload.menuId,
        viewType: "list",
        listSearch: this.search
      });
      this.env.services.action.navigate(url);
    }
    openRow(row) {
      const id = Number(row.id ?? 0);
      if (id <= 0) return;
      this.env.services.action.openRecord(
        this.props.payload.model,
        this.props.payload.actionId,
        this.props.payload.menuId,
        id,
        "form"
      );
    }
    template() {
      const p = this.props.payload;
      const cols = this.columns();
      const rows = this.rows();
      const fields = visibleFieldNames(cols);
      const reportActions = renderReportActions(p, fields);
      return html`
      <div class="sum-list-view">
        <div class="sum-view-toolbar">
          <div class="sum-view-toolbar-primary">
            ${renderNewButton(p)}
            <input
              type="search"
              class="sum-input sum-list-search"
              placeholder="Search…"
              value=${this.search}
              @keydown=${(ev) => ev.key === "Enter" && this.applySearch()}
              @input=${(ev) => {
        this.search = ev.target.value;
        this.bump?.();
      }}
            />
            <button type="button" class="sum-btn sum-btn--secondary" @click=${() => this.applySearch()}>Search</button>
          </div>
          ${reportActions ?? ""}
        </div>
        <table class="sum-list-table">
          <thead>
            <tr>${cols.map((c) => html`<th>${c.string ?? c.name}</th>`)}</tr>
          </thead>
          <tbody>
            ${rows.map(
        (row) => html`<tr class="sum-list-row" @click=${() => this.openRow(row)}>
                ${cols.map((c) => {
          const v = row[c.name];
          const display = row[`${c.name}_name`] ?? v;
          return html`<td>${String(display ?? "")}</td>`;
        })}
              </tr>`
      )}
          </tbody>
        </table>
      </div>
    `;
    }
  };

  // src/store/record.ts
  var SwcRecord = class {
    model;
    id;
    data;
    dirty = /* @__PURE__ */ new Set();
    constructor(model, id, data) {
      this.model = model;
      this.id = id;
      this.data = { ...data };
    }
    get(field) {
      return this.data[field];
    }
    set(field, value) {
      this.data[field] = value;
      this.dirty.add(field);
    }
    isDirty() {
      return this.dirty.size > 0;
    }
    dirtyValues() {
      const out = {};
      for (const k of this.dirty) {
        out[k] = this.data[k];
      }
      return out;
    }
    clearDirty() {
      this.dirty.clear();
    }
  };
  var RecordStore = class {
    rpc;
    constructor(rpc) {
      this.rpc = rpc;
    }
    fromPayload(model, id, data) {
      return new SwcRecord(model, id, data);
    }
    async save(rec) {
      if (rec.id <= 0) {
        const newId = await this.rpc.create(rec.model, rec.data);
        rec.clearDirty();
        return newId;
      }
      if (!rec.isDirty()) return rec.id;
      await this.rpc.write(rec.model, [rec.id], rec.dirtyValues());
      rec.clearDirty();
      return rec.id;
    }
    async unlink(rec) {
      if (rec.id <= 0) return;
      await this.rpc.unlink(rec.model, [rec.id]);
    }
    validate(rec, requiredFields) {
      for (const f of requiredFields) {
        const v = rec.get(f);
        if (v == null || v === "") {
          throw new SwcError(`Field ${f} is required`, "validation");
        }
      }
    }
  };

  // src/core/registry.ts
  var Registry = class {
    entries = /* @__PURE__ */ new Map();
    category(name) {
      if (!this.entries.has(name)) {
        this.entries.set(name, /* @__PURE__ */ new Map());
      }
      return new CategoryRegistry(this.entries.get(name));
    }
    get(category, key) {
      return this.entries.get(category)?.get(key);
    }
  };
  var CategoryRegistry = class {
    constructor(map) {
      this.map = map;
    }
    map;
    add(key, value) {
      this.map.set(key, value);
    }
    get(key) {
      return this.map.get(key);
    }
    keys() {
      return [...this.map.keys()];
    }
  };
  var registry = new Registry();

  // src/widgets/field-shell.ts
  function fieldInputId(field) {
    return `f-${field.name}`;
  }
  function isFullWidthField(field) {
    if (field.type === "text" || field.widget === "text") return true;
    if (field.type === "one2many" || field.widget === "one2many") return true;
    if (field.widget === "image") return true;
    return false;
  }
  function fieldWidgetClass(field, extra = []) {
    const parts = ["sum-field-widget"];
    if (isFullWidthField(field)) {
      parts.push("sum-field-widget--full");
    }
    if (field.type === "many2one" || field.widget === "many2one") {
      parts.push("sum-field-widget--many2one");
    }
    for (const mod of extra) {
      if (mod) parts.push(mod);
    }
    return parts.join(" ");
  }
  function fieldLabel(field, forId, row = false) {
    const label = field.string ?? field.name;
    const cls = row ? "sum-field-label sum-field-label--row" : "sum-field-label";
    if (forId) {
      return html`<label class=${cls} for=${forId}>${label}</label>`;
    }
    return html`<label class=${cls}>${label}</label>`;
  }
  function fieldControl(body, compact = false) {
    const cls = compact ? "sum-field-control sum-field-control--compact" : "sum-field-control";
    return html`<div class=${cls}>${body}</div>`;
  }
  function fieldPlaceholder(field) {
    return field.placeholder ?? field.string ?? field.name;
  }
  function fieldReadonlyValue(val, placeholder = "") {
    const hasValue = val.trim() !== "";
    const text = hasValue ? val : placeholder;
    const cls = hasValue ? "sum-field-value" : "sum-field-value sum-field-value--placeholder";
    return html`<div class=${cls}>${text}</div>`;
  }
  function fieldReadonlyInput(field, val, inputType = "text") {
    const placeholder = fieldPlaceholder(field);
    return html`<input
    type=${inputType}
    class="sum-field-input"
    name=${field.name}
    value=${val}
    placeholder=${placeholder}
    readonly
    tabindex="-1"
  />`;
  }
  function renderFieldShell(field, body, options = {}) {
    const showLabel = options.showLabel !== false;
    const labelFor = options.labelFor ?? fieldInputId(field);
    const useRow = options.layout === "row" || options.layout !== "stack" && !isFullWidthField(field) && !options.compact;
    const modifiers = [...options.modifiers ?? []];
    if (useRow) modifiers.push("sum-field-widget--row");
    const wrappedBody = fieldControl(body, options.compact === true);
    if (useRow) {
      return html`<div class=${fieldWidgetClass(field, modifiers)}>
      ${showLabel ? fieldLabel(field, labelFor, true) : ""}
      ${wrappedBody}
    </div>`;
    }
    return html`<div class=${fieldWidgetClass(field, modifiers)}>
    ${showLabel ? fieldLabel(field, labelFor) : ""}
    ${wrappedBody}
  </div>`;
  }

  // src/widgets/DefaultField.ts
  function inputTypeForField(field) {
    if (field.widget === "email") return "email";
    if (field.type === "integer" || field.type === "float" || field.type === "numeric") return "number";
    if (field.type === "date") return "date";
    if (field.type === "datetime") return "datetime-local";
    return "text";
  }
  function stepForField(field) {
    if (field.type === "integer") return "1";
    if (field.type === "float" || field.type === "numeric") return "any";
    return void 0;
  }
  function parseNumericValue(field, raw) {
    if (raw === "") return null;
    if (field.type === "integer") return Number.parseInt(raw, 10);
    if (field.type === "float" || field.type === "numeric") return Number.parseFloat(raw);
    return raw;
  }
  var DefaultField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const val = String(record.get(field.name) ?? "");
      const placeholder = fieldPlaceholder(field);
      const inputType = inputTypeForField(field);
      const step = stepForField(field);
      const id = fieldInputId(field);
      if (readonly || field.readonly) {
        return renderFieldShell(
          field,
          field.type === "integer" || field.type === "float" || field.type === "numeric" ? fieldReadonlyInput(field, val, "text") : fieldReadonlyValue(val, placeholder)
        );
      }
      return renderFieldShell(
        field,
        html`<input
        id=${id}
        type=${inputType}
        class="sum-field-input"
        name=${field.name}
        placeholder=${placeholder}
        value=${val}
        ${step ? html`step=${step}` : ""}
        @input=${(ev) => record.set(field.name, parseNumericValue(field, ev.target.value))}
      />`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/field-async.ts
  var AsyncFieldController = class {
    constructor(comp) {
      this.comp = comp;
    }
    comp;
    generation = 0;
    begin() {
      this.generation += 1;
      return this.generation;
    }
    cancel() {
      this.generation += 1;
    }
    refresh() {
      if (this.comp.el?.parentElement) {
        this.comp.patch();
      }
    }
    finish(gen) {
      if (gen !== this.generation) return;
      this.refresh();
    }
  };
  function recordDisplayName(record, fieldName) {
    const named = record.get(`${fieldName}_name`);
    if (named != null && named !== "") return String(named);
    const raw = record.get(fieldName);
    if (raw == null || raw === "") return "";
    return `#${raw}`;
  }

  // src/widgets/Many2OneField.ts
  var Many2OneField = class extends SwcComponent {
    suggestions = [];
    open = false;
    asyncCtrl = new AsyncFieldController(this);
    onWillUnmount() {
      this.asyncCtrl.cancel();
    }
    async search(q) {
      const gen = this.asyncCtrl.begin();
      const comodel = this.props.field.relation ?? this.props.field.options?.relation ?? "";
      if (!comodel) return;
      const domain = q ? [["name", "ilike", q]] : [];
      this.suggestions = await this.env.services.rpc.searchRead(comodel, domain, ["id", "name"], 20);
      this.open = true;
      this.asyncCtrl.finish(gen);
    }
    template() {
      const { field, record, readonly } = this.props;
      const display = record.get(`${field.name}_name`) ?? (record.get(field.name) ? `#${record.get(field.name)}` : "");
      const id = fieldInputId(field);
      const placeholder = fieldPlaceholder(field);
      if (readonly || field.readonly) {
        return renderFieldShell(field, fieldReadonlyValue(String(display), placeholder));
      }
      return renderFieldShell(
        field,
        html`<div class="sum-m2o-wrap">
        <input
          id=${id}
          class="sum-field-input"
          name=${field.name}
          placeholder=${placeholder}
          value=${String(display)}
          autocomplete="off"
          @input=${(ev) => void this.search(ev.target.value)}
        />
        ${this.open ? html`<ul class="sum-m2o-suggest">
              ${this.suggestions.map(
          (row) => html`<li>
                  <button
                    type="button"
                    class="sum-m2o-option"
                    @click=${() => {
            record.set(field.name, row.id);
            record.set(`${field.name}_name`, row.name);
            this.open = false;
            this.asyncCtrl.refresh();
          }}
                  >
                    ${String(row.name ?? row.id)}
                  </button>
                </li>`
        )}
            </ul>` : ""}
      </div>`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/StatusbarField.ts
  function isClickable(field) {
    const opt = field.options?.clickable;
    return opt !== "0" && opt !== "false";
  }
  var StatusbarField = class extends SwcComponent {
    stages = [];
    loaded = false;
    asyncCtrl = new AsyncFieldController(this);
    setup() {
      void this.loadStages();
    }
    onWillUnmount() {
      this.asyncCtrl.cancel();
    }
    async loadStages() {
      const gen = this.asyncCtrl.begin();
      const { field } = this.props;
      if (field.selection?.length) {
        this.stages = field.selection.map(([value, label]) => ({ id: value, label }));
        this.loaded = true;
        this.asyncCtrl.finish(gen);
        return;
      }
      const comodel = field.relation ?? field.options?.relation ?? "";
      if (!comodel) {
        const fallback = (field.options?.states ?? "draft,done").split(",").map((s) => s.trim()).filter(Boolean);
        this.stages = fallback.map((s) => ({ id: s, label: s }));
        this.loaded = true;
        this.asyncCtrl.finish(gen);
        return;
      }
      const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name", "sequence"], 200);
      rows.sort((a, b) => Number(a.sequence ?? 0) - Number(b.sequence ?? 0));
      this.stages = rows.map((row) => ({
        id: Number(row.id),
        label: String(row.name ?? row.id)
      }));
      this.loaded = true;
      this.asyncCtrl.finish(gen);
    }
    currentId() {
      const { field, record } = this.props;
      const raw = record.get(field.name);
      if (raw == null || raw === "") return "";
      return field.type === "many2one" || field.relation ? Number(raw) : String(raw);
    }
    template() {
      const { field, record, readonly } = this.props;
      const current = this.currentId();
      const clickable = isClickable(field) && !readonly && !field.readonly;
      return html`<div class="sum-statusbar-stages" role="group" aria-label=${field.string ?? field.name}>
      ${this.stages.map((stage) => {
        const active = stage.id === current || String(stage.id) === String(current);
        const stageClass = active ? "sum-statusbar-stage sum-statusbar-stage--current" : "sum-statusbar-stage";
        return html`<button type="button" class=${stageClass} disabled=${!clickable ? "disabled" : void 0} @click=${() => {
          if (!clickable) return;
          record.set(field.name, stage.id);
          if (field.type === "many2one" || field.relation) {
            record.set(`${field.name}_name`, stage.label);
          }
          this.asyncCtrl.refresh();
        }}>${stage.label}</button>`;
      })}
      ${!this.loaded ? html`<span class="sum-statusbar-chip">Loading…</span>` : ""}
    </div>`;
    }
  };

  // src/widgets/PriorityField.ts
  function priorityMode(field) {
    const mode = (field.options?.mode ?? field.options?.display ?? "stars").toLowerCase();
    return mode === "select" || mode === "dropdown" ? "select" : "stars";
  }
  function selectionOptions(field) {
    if (!field.selection?.length) {
      return [
        { value: "0", label: "Low" },
        { value: "1", label: "Medium" },
        { value: "2", label: "High" }
      ];
    }
    return field.selection.map(([value, label]) => ({ value, label }));
  }
  function currentValue(field, record) {
    const raw = record.get(field.name);
    if (raw == null || raw === "") return selectionOptions(field)[0]?.value ?? "0";
    return String(raw);
  }
  function numericLevel(value) {
    const n = Number.parseInt(value, 10);
    return Number.isNaN(n) ? 0 : Math.max(0, n);
  }
  function starCount(field) {
    const fromOpt = Number(field.options?.stars ?? field.options?.max ?? 0);
    if (fromOpt > 0) return Math.min(Math.max(fromOpt, 1), 5);
    const maxLevel = selectionOptions(field).length - 1;
    return Math.max(Math.min(maxLevel, 4), 3);
  }
  var PriorityField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const options = selectionOptions(field);
      const value = currentValue(field, record);
      const mode = priorityMode(field);
      if (readonly || field.readonly) {
        const label = options.find((o) => o.value === value)?.label ?? value;
        if (mode === "select") {
          return renderFieldShell(field, fieldReadonlyValue(label));
        }
        return renderFieldShell(field, this.renderStars(numericLevel(value), true));
      }
      if (mode === "select") {
        const id = fieldInputId(field);
        return renderFieldShell(
          field,
          html`<select
          id=${id}
          class="sum-field-select sum-priority-select"
          name=${field.name}
          @change=${(ev) => record.set(field.name, ev.target.value)}
        >
          ${options.map(
            (opt) => html`<option value=${opt.value} selected=${value === opt.value ? "selected" : ""}>
                ${opt.label}
              </option>`
          )}
        </select>`,
          { labelFor: id }
        );
      }
      return renderFieldShell(
        field,
        this.renderStars(numericLevel(value), false, (level) => {
          record.set(field.name, String(level));
        })
      );
    }
    starButtons(level, disabled, onPick) {
      const { field } = this.props;
      const options = selectionOptions(field);
      const count = starCount(field);
      const capped = Math.min(level, count);
      const out = [];
      for (let i = 0; i < count; i += 1) {
        const starIndex = i + 1;
        const filled = starIndex <= capped;
        const opt = options[Math.min(starIndex, options.length - 1)];
        const click = () => {
          if (disabled) return;
          const next = capped === starIndex ? starIndex - 1 : starIndex;
          onPick?.(Math.max(0, next));
        };
        if (filled) {
          out.push(html`<button type="button" class="sum-priority-star sum-priority-star--on" disabled=${disabled ? "disabled" : void 0} title=${opt?.label ?? `Level ${starIndex}`} aria-label=${opt?.label ?? `Priority ${starIndex}`} @click=${click}>★</button>`);
        } else {
          out.push(html`<button type="button" class="sum-priority-star" disabled=${disabled ? "disabled" : void 0} title=${opt?.label ?? `Level ${starIndex}`} aria-label=${opt?.label ?? `Priority ${starIndex}`} @click=${click}>★</button>`);
        }
      }
      return out;
    }
    renderStars(level, disabled, onPick) {
      const { field } = this.props;
      return html`<div class="sum-priority-stars" role="group" aria-label=${field.string ?? field.name}>
      ${this.starButtons(level, disabled, onPick)}
    </div>`;
    }
  };

  // src/widgets/BooleanField.ts
  function isChecked(val) {
    return val === true || val === 1 || val === "1" || val === "true";
  }
  var BooleanField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const checked = isChecked(record.get(field.name));
      const id = fieldInputId(field);
      if (readonly || field.readonly) {
        return renderFieldShell(field, fieldReadonlyValue(checked ? "Yes" : "No"));
      }
      return renderFieldShell(
        field,
        html`<input
        id=${id}
        type="checkbox"
        class="sum-field-input"
        name=${field.name}
        checked=${checked ? "checked" : ""}
        @change=${(ev) => record.set(field.name, ev.target.checked)}
      />`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/TextareaField.ts
  var TextareaField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const val = String(record.get(field.name) ?? "");
      const placeholder = fieldPlaceholder(field);
      const id = fieldInputId(field);
      if (readonly || field.readonly) {
        return renderFieldShell(field, fieldReadonlyValue(val, placeholder));
      }
      return renderFieldShell(
        field,
        html`<textarea
        id=${id}
        class="sum-field-textarea"
        name=${field.name}
        placeholder=${placeholder}
        rows="5"
        @input=${(ev) => record.set(field.name, ev.target.value)}
      >${val}</textarea>`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/SelectionField.ts
  var SelectionField = class extends SwcComponent {
    options = [];
    loaded = false;
    asyncCtrl = new AsyncFieldController(this);
    setup() {
      void this.loadOptions();
    }
    onWillUnmount() {
      this.asyncCtrl.cancel();
    }
    async loadOptions() {
      const gen = this.asyncCtrl.begin();
      const { field, readonly } = this.props;
      if (field.selection?.length) {
        this.options = field.selection.map(([value, label]) => ({ value, label }));
        this.loaded = true;
        this.asyncCtrl.finish(gen);
        return;
      }
      if (readonly || field.readonly) {
        this.loaded = true;
        this.asyncCtrl.finish(gen);
        return;
      }
      const comodel = field.relation ?? field.options?.relation ?? "";
      if (!comodel) {
        this.loaded = true;
        this.asyncCtrl.finish(gen);
        return;
      }
      const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name"], 200);
      this.options = (rows ?? []).map((row) => ({
        value: String(row.id ?? ""),
        label: String(row.name ?? row.id ?? "")
      }));
      this.loaded = true;
      this.asyncCtrl.finish(gen);
    }
    displayValue() {
      const { field, record } = this.props;
      const raw = record.get(field.name);
      const id = raw == null || raw === "" ? "" : String(raw);
      if (!id) return "";
      const named = record.get(`${field.name}_name`);
      if (named) return String(named);
      const match = this.options.find((o) => o.value === id);
      return match?.label ?? recordDisplayName(record, field.name);
    }
    template() {
      const { field, record, readonly } = this.props;
      const current = record.get(field.name);
      const currentVal = current == null || current === "" ? "" : String(current);
      const id = fieldInputId(field);
      const placeholder = fieldPlaceholder(field);
      if (readonly || field.readonly) {
        return renderFieldShell(field, fieldReadonlyValue(this.displayValue(), placeholder));
      }
      return renderFieldShell(
        field,
        html`<select
        id=${id}
        class="sum-field-input sum-field-select"
        name=${field.name}
        @change=${(ev) => {
          const val = ev.target.value;
          const opt = this.options.find((o) => o.value === val);
          record.set(field.name, val ? Number(val) || val : null);
          if (opt) record.set(`${field.name}_name`, opt.label);
          this.asyncCtrl.refresh();
        }}
      >
        <option value="" disabled=${currentVal !== "" ? "disabled" : false} selected=${currentVal === "" ? "selected" : false}>${placeholder}</option>
        ${this.options.map(
          (opt) => html`<option value=${opt.value} selected=${opt.value === currentVal ? "selected" : ""}>
              ${opt.label}
            </option>`
        )}
      </select>
      ${!this.loaded ? html`<span class="sum-field-hint">Loading…</span>` : ""}`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/PhoneField.ts
  var PhoneField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const val = String(record.get(field.name) ?? "");
      const placeholder = fieldPlaceholder(field);
      const id = fieldInputId(field);
      if (readonly || field.readonly) {
        return renderFieldShell(field, fieldReadonlyValue(val, placeholder));
      }
      return renderFieldShell(
        field,
        html`<input
        id=${id}
        type="tel"
        class="sum-field-input sum-field-phone"
        name=${field.name}
        placeholder=${placeholder}
        value=${val}
        @input=${(ev) => record.set(field.name, ev.target.value)}
      />`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/BooleanRadioField.ts
  function isChecked2(val) {
    return val === true || val === 1 || val === "1" || val === "true";
  }
  var BooleanRadioField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const checked = isChecked2(record.get(field.name));
      const name = field.name;
      return renderFieldShell(
        field,
        html`<div class="sum-field-radio-group" role="radiogroup">
        <label class="sum-field-radio">
          <input
            type="radio"
            name=${name}
            value="1"
            checked=${checked ? "checked" : ""}
            disabled=${readonly || field.readonly ? "disabled" : void 0}
            @change=${() => !readonly && record.set(field.name, true)}
          />
          Yes
        </label>
        <label class="sum-field-radio">
          <input
            type="radio"
            name=${name}
            value="0"
            checked=${!checked ? "checked" : ""}
            disabled=${readonly || field.readonly ? "disabled" : void 0}
            @change=${() => !readonly && record.set(field.name, false)}
          />
          No
        </label>
      </div>`
      );
    }
  };

  // src/widgets/BooleanToggleField.ts
  function isChecked3(val) {
    return val === true || val === 1 || val === "1" || val === "true";
  }
  var BooleanToggleField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const checked = isChecked3(record.get(field.name));
      const id = fieldInputId(field);
      return renderFieldShell(
        field,
        html`<label class="sum-field-toggle" for=${id}>
        <input
          id=${id}
          type="checkbox"
          class="sum-field-input"
          name=${field.name}
          checked=${checked ? "checked" : ""}
          disabled=${readonly || field.readonly ? "disabled" : void 0}
          @change=${(ev) => record.set(field.name, ev.target.checked)}
        />
        <span>${checked ? "On" : "Off"}</span>
      </label>`,
        { showLabel: true, labelFor: id }
      );
    }
  };

  // src/widgets/Many2ManyTagsField.ts
  function tagIds(record, fieldName) {
    const raw = record.get(fieldName);
    if (!Array.isArray(raw)) return [];
    return raw.map((v) => Number(v)).filter((n) => !Number.isNaN(n));
  }
  function tagNamesFromRecord(record, fieldName) {
    const out = /* @__PURE__ */ new Map();
    const raw = record.get(`${fieldName}_names`);
    if (Array.isArray(raw)) {
      for (const item of raw) {
        if (item && typeof item === "object") {
          const row = item;
          const id = Number(row.id);
          if (!Number.isNaN(id)) out.set(id, String(row.name ?? id));
        }
      }
    }
    return out;
  }
  var Many2ManyTagsField = class extends SwcComponent {
    catalog = [];
    loaded = false;
    asyncCtrl = new AsyncFieldController(this);
    setup() {
      void this.loadCatalog();
    }
    onWillUnmount() {
      this.asyncCtrl.cancel();
    }
    async loadCatalog() {
      const gen = this.asyncCtrl.begin();
      const { field, readonly } = this.props;
      if (readonly || field.readonly) {
        this.loaded = true;
        this.asyncCtrl.finish(gen);
        return;
      }
      const comodel = field.relation ?? field.options?.relation ?? "";
      if (!comodel) {
        this.loaded = true;
        this.asyncCtrl.finish(gen);
        return;
      }
      const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name"], 500);
      this.catalog = rows.map((row) => ({ id: Number(row.id), name: String(row.name ?? row.id) }));
      this.loaded = true;
      this.asyncCtrl.finish(gen);
    }
    selectedTags() {
      const { field, record } = this.props;
      const ids = tagIds(record, field.name);
      const names = tagNamesFromRecord(record, field.name);
      return ids.map((id) => {
        const fromCatalog = this.catalog.find((t) => t.id === id);
        if (fromCatalog) return fromCatalog;
        const fromRecord = names.get(id);
        if (fromRecord) return { id, name: fromRecord };
        return { id, name: `#${id}` };
      });
    }
    setIds(ids) {
      this.props.record.set(this.props.field.name, ids);
      this.asyncCtrl.refresh();
    }
    template() {
      const { field, readonly } = this.props;
      const selected = this.selectedTags();
      const selectedSet = new Set(selected.map((t) => t.id));
      if (readonly || field.readonly) {
        return renderFieldShell(
          field,
          html`<div class="sum-multi-select-tags sum-multi-select-tags--readonly sum-field-tags">
          ${selected.map((tag) => html`<span class="sum-multi-select-tag"><span class="sum-multi-select-tag-label">${tag.name}</span></span>`)}
        </div>`
        );
      }
      return renderFieldShell(
        field,
        html`<div class="sum-multi-select-box">
        <div class="sum-multi-select-tags sum-field-tags">
          ${selected.map(
          (tag) => html`<span class="sum-multi-select-tag">
                <span class="sum-multi-select-tag-label">${tag.name}</span>
                <button type="button" class="sum-multi-select-remove" aria-label="Remove" @click=${() => this.setIds(selected.filter((t) => t.id !== tag.id).map((t) => t.id))}>×</button>
              </span>`
        )}
        </div>
        <select
          class="sum-multi-select-add sum-field-select"
          @change=${(ev) => {
          const val = Number(ev.target.value);
          ev.target.value = "";
          if (!val || selectedSet.has(val)) return;
          this.setIds([...selected.map((t) => t.id), val]);
        }}
        >
          <option value="">Add tag…</option>
          ${this.catalog.filter((t) => !selectedSet.has(t.id)).map((t) => html`<option value=${String(t.id)}>${t.name}</option>`)}
        </select>
        ${!this.loaded ? html`<span class="sum-field-hint">Loading…</span>` : ""}
      </div>`
      );
    }
  };

  // src/widgets/One2ManyField.ts
  var tempLineId = -1;
  function nextTempId() {
    tempLineId -= 1;
    return tempLineId;
  }
  function inverseFieldName(parentModel) {
    const part = parentModel.split(".").pop() ?? "parent";
    return `${part}_id`;
  }
  function columnsForField(field) {
    return field.subview?.fields ?? [];
  }
  function columnNames(cols) {
    return cols.map((c) => c.name);
  }
  function parseCellValue(col, raw) {
    if (raw === "") return null;
    if (col.type === "integer") return Number.parseInt(raw, 10);
    if (col.type === "float" || col.type === "numeric") return Number.parseFloat(raw);
    if (col.type === "boolean") return raw === "true" || raw === "1";
    return raw;
  }
  function displayCellValue(col, line) {
    const raw = line[col.name];
    if (raw == null) return "";
    const named = line[`${col.name}_name`];
    if (named != null && String(named) !== "") return String(named);
    if (col.type === "boolean") {
      return raw === true || raw === 1 || raw === "1" || raw === "true" ? "Yes" : "No";
    }
    return String(raw);
  }
  var One2ManyField = class extends SwcComponent {
    lines = [];
    loaded = false;
    saving = false;
    asyncCtrl = new AsyncFieldController(this);
    writeTimers = /* @__PURE__ */ new Map();
    setup() {
      void this.loadLines();
    }
    onWillUnmount() {
      this.asyncCtrl.cancel();
      for (const t of this.writeTimers.values()) clearTimeout(t);
      this.writeTimers.clear();
    }
    comodel() {
      const { field } = this.props;
      return field.relation ?? field.options?.relation ?? "";
    }
    inverse() {
      const { field, record } = this.props;
      return field.options?.inverse ?? inverseFieldName(record.model);
    }
    editable() {
      const { field, record, readonly } = this.props;
      if (readonly || field.readonly) return false;
      if (record.id <= 0) return false;
      const mode = field.subview?.editable ?? "bottom";
      return mode === "bottom" || mode === "top";
    }
    async loadLines() {
      const gen = this.asyncCtrl.begin();
      const { field, record } = this.props;
      const comodel = this.comodel();
      const cols = columnsForField(field);
      if (!comodel || record.id <= 0 || cols.length === 0) {
        this.loaded = true;
        this.asyncCtrl.finish(gen);
        return;
      }
      const inv = this.inverse();
      const names = ["id", ...columnNames(cols)];
      const rows = await this.env.services.rpc.searchRead(
        comodel,
        [[inv, "=", record.id]],
        names,
        200
      );
      this.lines = (rows ?? []).map((row) => ({
        id: Number(row.id ?? 0),
        data: { ...row }
      }));
      this.loaded = true;
      this.asyncCtrl.finish(gen);
    }
    lineById(id) {
      return this.lines.find((l) => l.id === id);
    }
    scheduleWrite(lineId, col, value) {
      const key = `${lineId}:${col.name}`;
      const prev = this.writeTimers.get(key);
      if (prev) clearTimeout(prev);
      this.writeTimers.set(
        key,
        setTimeout(() => {
          this.writeTimers.delete(key);
          void this.persistCell(lineId, col, value);
        }, 350)
      );
    }
    async persistCell(lineId, col, value) {
      if (lineId <= 0) return;
      const comodel = this.comodel();
      if (!comodel) return;
      this.saving = true;
      this.asyncCtrl.refresh();
      try {
        await this.env.services.rpc.write(comodel, [lineId], { [col.name]: value });
        const line = this.lineById(lineId);
        if (line) line.data[col.name] = value;
      } finally {
        this.saving = false;
        this.asyncCtrl.refresh();
      }
    }
    async createLine(lineId, col, value) {
      const { record } = this.props;
      const comodel = this.comodel();
      const line = this.lineById(lineId);
      if (!comodel || !line || line.id > 0) return;
      line.data[col.name] = value;
      this.saving = true;
      this.asyncCtrl.refresh();
      try {
        const vals = { ...line.data, [this.inverse()]: record.id };
        delete vals.id;
        const newId = await this.env.services.rpc.create(comodel, vals);
        line.id = newId;
        line.data.id = newId;
      } finally {
        this.saving = false;
        this.asyncCtrl.refresh();
      }
    }
    onCellInput(lineId, col, raw) {
      const value = typeof raw === "boolean" ? raw : parseCellValue(col, String(raw ?? ""));
      const line = this.lineById(lineId);
      if (!line) return;
      line.data[col.name] = value;
      if (line.id <= 0) {
        void this.createLine(lineId, col, value);
        return;
      }
      this.scheduleWrite(line.id, col, value);
    }
    addRow() {
      const id = nextTempId();
      this.lines = [...this.lines, { id, data: {} }];
      this.asyncCtrl.refresh();
    }
    async deleteRow(lineId) {
      const comodel = this.comodel();
      const line = this.lineById(lineId);
      if (!line) return;
      if (line.id > 0 && comodel) {
        this.saving = true;
        this.asyncCtrl.refresh();
        try {
          await this.env.services.rpc.unlink(comodel, [line.id]);
        } finally {
          this.saving = false;
        }
      }
      this.lines = this.lines.filter((l) => l.id !== lineId);
      this.asyncCtrl.refresh();
    }
    renderCellEditor(col, line) {
      const val = String(line.data[col.name] ?? "");
      const readonly = !this.editable();
      if (readonly) {
        return html`<span>${displayCellValue(col, line.data)}</span>`;
      }
      if (col.type === "boolean") {
        const checked = line.data[col.name] === true || line.data[col.name] === 1;
        return fieldControl(
          html`<input
          type="checkbox"
          class="sum-field-input"
          checked=${checked ? "checked" : ""}
          @change=${(ev) => this.onCellInput(line.id, col, ev.target.checked)}
        />`,
          true
        );
      }
      if (col.selection?.length) {
        return fieldControl(
          html`<select
          class="sum-field-select"
          @change=${(ev) => this.onCellInput(line.id, col, ev.target.value)}
        >
          <option value="">—</option>
          ${col.selection.map(
            ([v, label]) => html`<option value=${v} selected=${val === v ? "selected" : ""}>${label}</option>`
          )}
        </select>`,
          true
        );
      }
      const inputType = col.type === "integer" || col.type === "float" || col.type === "numeric" ? "number" : col.type === "date" ? "date" : "text";
      return fieldControl(
        html`<input
        type=${inputType}
        class="sum-field-input"
        value=${val}
        @input=${(ev) => this.onCellInput(line.id, col, ev.target.value)}
      />`,
        true
      );
    }
    renderLineRow(line, cols, canEdit) {
      const cells = cols.map(
        (col) => html`<td>${this.renderCellEditor(col, line)}</td>`
      );
      if (canEdit) {
        cells.push(html`<td class="sum-o2m-col-actions"><button type="button" .sum-o2m-delete-btn data-line-id=${String(line.id)} title="Remove line">×</button></td>`);
      }
      return html`<tr class="sum-o2m-row">${cells}</tr>`;
    }
    onTableClick(ev) {
      const btn = ev.target.closest(".sum-o2m-delete-btn");
      if (!btn) return;
      const id = Number(btn.getAttribute("data-line-id"));
      if (!Number.isFinite(id)) return;
      void this.deleteRow(id);
    }
    template() {
      const { field, record } = this.props;
      const label = field.string ?? field.name;
      const cols = columnsForField(field);
      const canEdit = this.editable();
      const emptyMsg = !this.loaded ? "Loading\u2026" : record.id <= 0 ? "Save the record before adding lines." : cols.length === 0 ? "No columns configured." : "No lines";
      return renderFieldShell(
        field,
        html`<div class="sum-o2m-table-wrap">
        <div class="sum-o2m-title">${label}${this.saving ? " (saving\u2026)" : ""}</div>
        <table class="sum-o2m-table">
          <thead>
            <tr>
              ${cols.map((col) => html`<th>${col.string ?? col.name}</th>`)}
              ${canEdit ? html`<th class="sum-o2m-col-actions"></th>` : ""}
            </tr>
          </thead>
          <tbody @click=${(ev) => this.onTableClick(ev)}>
            ${this.lines.length === 0 ? html`<tr>
                  <td colspan=${String(cols.length + (canEdit ? 1 : 0))}>${emptyMsg}</td>
                </tr>` : this.lines.map((line) => this.renderLineRow(line, cols, canEdit))}
          </tbody>
        </table>
        ${canEdit && cols.length > 0 ? html`<button type="button" class="sum-o2m-add-row" @click=${() => this.addRow()}>
              + Add a line
            </button>` : ""}
        ${!canEdit && record.id <= 0 && !this.props.readonly ? html`<p class="sum-o2m-hint">Save the parent record before editing lines.</p>` : ""}
      </div>`,
        { layout: "stack", showLabel: false }
      );
    }
  };

  // src/widgets/DateField.ts
  function isDateTime(field) {
    return field.type === "datetime" || field.widget === "datetime";
  }
  function toNativeValue(field, raw) {
    const text = String(raw ?? "").trim();
    if (!text) return "";
    if (isDateTime(field)) {
      const d = new Date(text);
      if (Number.isNaN(d.getTime())) return text.slice(0, 16);
      const pad = (n) => String(n).padStart(2, "0");
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
    }
    return text.slice(0, 10);
  }
  function formatDisplay(field, raw) {
    const native = toNativeValue(field, raw);
    if (!native) return "";
    if (isDateTime(field)) {
      const d2 = new Date(native);
      if (Number.isNaN(d2.getTime())) return native;
      return d2.toLocaleString(void 0, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit"
      });
    }
    const d = /* @__PURE__ */ new Date(`${native}T00:00:00`);
    if (Number.isNaN(d.getTime())) return native;
    return d.toLocaleDateString(void 0, { year: "numeric", month: "short", day: "numeric" });
  }
  function todayNative(field) {
    const d = /* @__PURE__ */ new Date();
    const pad = (n) => String(n).padStart(2, "0");
    if (isDateTime(field)) {
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
    }
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  }
  function closeDetails(ev) {
    const details = ev.currentTarget?.closest("details.sum-date-field");
    if (details instanceof HTMLDetailsElement) details.open = false;
  }
  var DateField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const raw = record.get(field.name);
      const native = toNativeValue(field, raw);
      const display = formatDisplay(field, raw);
      const placeholder = fieldPlaceholder(field);
      const id = fieldInputId(field);
      const inputType = isDateTime(field) ? "datetime-local" : "date";
      if (readonly || field.readonly) {
        return renderFieldShell(field, fieldReadonlyValue(display, placeholder));
      }
      return renderFieldShell(
        field,
        html`<details class="sum-date-field">
        <summary class="sum-date-field-trigger">
          <span class=${display ? "sum-date-field-value" : "sum-date-field-value sum-date-field-value--placeholder"}>
            ${display || placeholder}
          </span>
          <span class="sum-date-field-icon" aria-hidden="true">📅</span>
        </summary>
        <input type="hidden" id=${id} name=${field.name} value=${native} />
        <div class="sum-date-popover" role="dialog" aria-label=${placeholder}>
          <div class="sum-date-popover-header">${field.string ?? field.name}</div>
          <input
            type=${inputType}
            class="sum-date-popover-input"
            value=${native}
            @input=${(ev) => {
          record.set(field.name, ev.target.value || null);
          this.patch();
        }}
            @change=${(ev) => {
          record.set(field.name, ev.target.value || null);
          this.patch();
        }}
          />
          <div class="sum-date-popover-actions">
            <button
              type="button"
              class="sum-date-popover-btn"
              @click=${(ev) => {
          record.set(field.name, todayNative(field));
          this.patch();
          closeDetails(ev);
        }}
            >
              Today
            </button>
            <button
              type="button"
              class="sum-date-popover-btn"
              @click=${(ev) => {
          record.set(field.name, null);
          this.patch();
          closeDetails(ev);
        }}
            >
              Clear
            </button>
            <button type="button" class="sum-date-popover-btn sum-date-popover-btn--primary" @click=${closeDetails}>
              Done
            </button>
          </div>
        </div>
      </details>`,
        { labelFor: id }
      );
    }
  };

  // src/widgets/ImageField.ts
  var ImageField = class extends SwcComponent {
    template() {
      const { field, record, readonly } = this.props;
      const image = String(record.get(field.name) ?? "");
      const hasImage = image.length > 0;
      return renderFieldShell(
        field,
        html`<div data-sum-avatar>
        ${hasImage ? html`<div class="sum-image-thumb"><img class="sum-image-thumb-img" src=${image} alt="" /></div>` : html`<div class="sum-image-thumb sum-image-thumb--empty">No image</div>`}
        ${readonly || field.readonly ? html`<input type="hidden" data-sum-image-value name=${field.name} value=${image} />` : html`<label class="sum-form-avatar-upload">
              Upload
              <input type="file" accept="image/*" />
              <input
                type="hidden"
                data-sum-image-value
                name=${field.name}
                value=${image}
                @input=${(ev) => record.set(field.name, ev.target.value)}
              />
            </label>`}
      </div>`,
        { modifiers: ["sum-field-widget--image"] }
      );
    }
  };

  // src/widgets/registry.ts
  function registerDefaultWidgets() {
    const fields = registry.category("fields");
    const add = (key, Ctor) => fields.add(key, Ctor);
    add("default", DefaultField);
    add("char", DefaultField);
    add("email", DefaultField);
    add("integer", DefaultField);
    add("float", DefaultField);
    add("numeric", DefaultField);
    add("date", DateField);
    add("datetime", DateField);
    add("json", TextareaField);
    add("many2one", Many2OneField);
    add("one2many", One2ManyField);
    add("many2many", Many2ManyTagsField);
    add("selection", SelectionField);
    add("boolean", BooleanField);
    add("text", TextareaField);
    add("statusbar", StatusbarField);
    add("priority", PriorityField);
    add("phone", PhoneField);
    add("radio", BooleanRadioField);
    add("boolean_toggle", BooleanToggleField);
    add("many2many_tags", Many2ManyTagsField);
    add("image", ImageField);
  }
  function resolveFieldWidget(field) {
    if (field.widget === "many2many_tags") return "many2many_tags";
    if (field.widget === "boolean_toggle") return "boolean_toggle";
    if (field.widget === "radio") return "radio";
    if (field.widget === "phone") return "phone";
    if (field.widget === "image") return "image";
    if (field.widget === "selection") return "selection";
    if (field.widget === "email") return "email";
    if (field.widget === "statusbar") return "statusbar";
    if (field.widget === "priority") return "priority";
    if (field.type === "boolean" && field.widget === "radio") return "radio";
    if (field.type === "boolean") return "boolean";
    if (field.type === "text") return "text";
    if (field.type === "many2one") return "many2one";
    if (field.type === "one2many") return "one2many";
    if (field.type === "many2many") return "many2many_tags";
    if (field.type === "selection") return "selection";
    if (field.type === "date") return "date";
    if (field.type === "datetime") return "datetime";
    if (field.type === "integer" || field.type === "float" || field.type === "numeric") {
      return field.type;
    }
    return field.widget ?? field.type ?? "default";
  }
  function renderField(env, field, record, readonly) {
    const key = resolveFieldWidget(field);
    const Ctor = registry.get("fields", key) ?? registry.get("fields", "default");
    const comp = new Ctor({ field, record, readonly }, env);
    comp.setup?.();
    return comp.render();
  }

  // src/views/form-sheet.ts
  function renderFields(rf, fields, record, readonly) {
    return visibleFields(fields).map((f) => rf(f, record, readonly));
  }
  function collectDivFields(div) {
    const out = [...div.fields ?? [], ...div.h1Fields ?? []];
    for (const nested of div.divs ?? []) {
      out.push(...collectDivFields(nested));
    }
    return out;
  }
  function collectFormFields(sheet, headerFields = []) {
    const out = [...headerFields];
    if (!sheet) return out.filter((f) => !f.invisible);
    out.push(...sheet.fields ?? []);
    for (const div of sheet.divs ?? []) {
      out.push(...collectDivFields(div));
    }
    for (const g of sheet.groups ?? []) {
      out.push(...collectGroupFields(g));
    }
    for (const nb of sheet.notebook ?? []) {
      for (const pg of nb.pages ?? []) {
        out.push(...pg.fields ?? []);
        for (const g of pg.groups ?? []) {
          out.push(...collectGroupFields(g));
        }
      }
    }
    return out.filter((f) => !f.invisible);
  }
  function collectGroupFields(group) {
    const out = [...group.fields ?? []];
    for (const nested of group.groups ?? []) {
      out.push(...collectGroupFields(nested));
    }
    return out;
  }
  function visibleFields(fields) {
    return fields.filter((f) => !f.invisible);
  }
  function renderSeparators(separators = []) {
    if (separators.length === 0) return html``;
    return html`${separators.map(
      (sep) => sep.string ? html`<div class="sum-separator--title">${sep.string}</div>` : html`<hr class="sum-separator--rule" />`
    )}`;
  }
  function renderLabels(labels = []) {
    if (labels.length === 0) return html``;
    return html`${labels.map((lab) => {
      const text = lab.string ?? "";
      if (lab.for) {
        return html`<label class="sum-form-label sum-form-label--hint" for=${`f-${lab.for}`}>${text}</label>`;
      }
      return html`<div class="sum-form-label sum-form-label--hint">${text}</div>`;
    })}`;
  }
  function initialsFromName(name) {
    const parts = name.trim().split(/\s+/).filter(Boolean);
    if (parts.length === 0) return "?";
    if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }
  function renderHeroField(field, record, readonly) {
    const val = String(record.get(field.name) ?? "");
    const placeholder = fieldPlaceholder(field);
    const hasValue = val.trim() !== "";
    if (readonly || field.readonly) {
      const text = hasValue ? val : placeholder;
      const cls = hasValue ? "sum-form-hero-input sum-form-hero-input--bold" : "sum-form-hero-input sum-form-hero-input--bold sum-form-hero-input--placeholder";
      return html`<h1><div class=${cls}>${text}</div></h1>`;
    }
    return html`<h1>
    <input
      class="sum-form-hero-input sum-form-hero-input--bold"
      name=${field.name}
      placeholder=${placeholder}
      value=${val}
      aria-label=${placeholder}
      @input=${(ev) => record.set(field.name, ev.target.value)}
    />
  </h1>`;
  }
  function renderContactItem(field, record, readonly) {
    const val = String(record.get(field.name) ?? "");
    const label = field.string ?? field.name;
    const placeholder = fieldPlaceholder(field);
    const inputType = field.widget === "email" ? "email" : "text";
    if (readonly || field.readonly) {
      const text = val.trim() !== "" ? val : placeholder;
      const cls = val.trim() !== "" ? "sum-form-inline-input" : "sum-form-inline-input sum-form-inline-input--placeholder";
      return html`<div class="sum-form-contact-item">
      <label class="sum-field-label">${label}</label>
      <div class=${cls}>${text}</div>
    </div>`;
    }
    return html`<div class="sum-form-contact-item">
    <label class="sum-field-label">${label}</label>
    <input
      type=${inputType}
      class="sum-form-inline-input"
      name=${field.name}
      placeholder=${placeholder}
      value=${val}
      @input=${(ev) => record.set(field.name, ev.target.value)}
    />
  </div>`;
  }
  function renderAvatar(record, readonly) {
    const image = String(record.get("image") ?? "");
    const name = String(record.get("name") ?? "");
    const hasImage = image.length > 0;
    const initials2 = initialsFromName(name);
    return html`<div class="sum-form-avatar sum-form-avatar--compact" data-sum-avatar>
    <div class="sum-form-avatar-box sum-form-avatar-box--circle">
      ${hasImage ? html`<img
            .sum-form-avatar-img
            .sum-form-avatar-img--visible
            class=${image.includes("data:") ? "sum-form-avatar-img--cropped" : ""}
            src=${image}
            alt=""
          />` : html`<span class="sum-form-avatar-initials">${initials2}</span>`}
    </div>
    ${readonly ? "" : html`<div class="sum-form-avatar-actions">
          <input
            type="hidden"
            name="image"
            data-sum-avatar-value
            value=${image}
            @input=${(ev) => record.set("image", ev.target.value)}
          />
          <label class="sum-form-avatar-upload">
            Upload
            <input type="file" accept="image/*" />
          </label>
        </div>`}
  </div>`;
  }
  function renderTitleBody(rf, div, record, readonly) {
    const h1Fields = visibleFields(div.h1Fields ?? []);
    const contactDiv = (div.divs ?? []).find((d) => (d.class ?? "").includes("sum-title-contact-row"));
    const contactFields = visibleFields(contactDiv?.fields ?? []);
    return html`<div class="sum-form-title-body sum-form-title-body--main">
    ${h1Fields.length > 0 ? renderHeroField(h1Fields[0], record, readonly) : ""}
    ${contactFields.length > 0 ? html`<div class="sum-title-contact-row">
          ${contactFields.map((f) => renderContactItem(f, record, readonly))}
        </div>` : ""}
    ${h1Fields.length === 0 && contactFields.length === 0 ? renderFields(rf, div.fields ?? [], record, readonly) : ""}
  </div>`;
  }
  function renderTitleDiv(rf, div, record, readonly, hasImageField) {
    const cls = div.class ?? "";
    if (cls.includes("sum_button_box")) {
      return html`<div class="sum-form-button-box ${cls}"></div>`;
    }
    const isTitle = cls.includes("sum_title");
    if (!isTitle) {
      return html`<div class=${cls}>${renderFields(rf, div.fields ?? [], record, readonly)}</div>`;
    }
    const h1Fields = visibleFields(div.h1Fields ?? []);
    const legacySingle = h1Fields.length === 0 && visibleFields(div.fields ?? []).length === 1;
    const titleField2 = h1Fields[0] ?? (legacySingle ? visibleFields(div.fields ?? [])[0] : void 0);
    if (hasImageField) {
      return html`<div class="sum-form-split-layout sum-form-split-layout--compact" data-sum-form-split>
      <aside class="sum-form-split-left sum-form-split-left--avatar">${renderAvatar(record, readonly)}</aside>
      <div class="sum-form-split-main">${renderTitleBody(rf, div, record, readonly)}</div>
    </div>`;
    }
    if (titleField2) {
      return html`<div class="sum-form-title-row sum-form-title-row--sheet">
      ${renderTitleBody(rf, div, record, readonly)}
    </div>`;
    }
    return html`<div class="sum-form-title-row sum-form-title-row--sheet">
    ${renderTitleBody(rf, div, record, readonly)}
  </div>`;
  }
  function outerGroupMaxCols(group, childCount) {
    if (group.col && group.col > 0) return group.col;
    return Math.max(childCount, 1);
  }
  function childGroupColspan(group) {
    return group.colspan && group.colspan > 0 ? group.colspan : 1;
  }
  function gridSpan12(maxCols, colspan) {
    const cols = Math.max(maxCols, 1);
    const span = Math.max(colspan, 1);
    return Math.min(12, Math.max(1, Math.round(span * 12 / cols)));
  }
  function packGroupRows(parent, nested) {
    const maxCols = outerGroupMaxCols(parent, nested.length);
    const rows = [];
    let current = [];
    let used = 0;
    for (const child of nested) {
      const colspan = childGroupColspan(child);
      if (used + colspan > maxCols && current.length > 0) {
        rows.push(current);
        current = [];
        used = 0;
      }
      current.push({ group: child, gridSpan: gridSpan12(maxCols, colspan) });
      used += colspan;
    }
    if (current.length > 0) rows.push(current);
    return rows;
  }
  function groupClassNames(group, ctx, plain) {
    const parts = ["sum-form-group"];
    if (plain || !group.string) {
      parts.push("sum-form-group--plain");
    } else if (ctx === "row" || ctx === "inner") {
      parts.push("sum-form-group--col");
    } else {
      parts.push("sum-form-group--full");
    }
    if ((group.fields ?? []).length > 0) {
      parts.push("sum-form-group--row-layout");
    }
    return parts.join(" ");
  }
  function renderGroup(rf, group, record, readonly, ctx = "sheet", plain = false) {
    const nested = group.groups ?? [];
    const fields = group.fields ?? [];
    const hasNested = nested.length > 0;
    if (hasNested && fields.length === 0) {
      const rows = packGroupRows(group, nested);
      return html`<div class="sum-form-group-outer sum-field-region--sheet">
      ${rows.map(
        (row) => html`<div class="sum-form-group-row">
          ${row.map(
          (item) => html`<div class="sum-form-group-span" style=${`--sum-group-span:${item.gridSpan}`}>
              ${renderGroup(rf, item.group, record, readonly, "row")}
            </div>`
        )}
        </div>`
      )}
    </div>`;
    }
    const innerCols = group.col && group.col > 0 ? group.col : 0;
    const innerColsClass = innerCols > 0 ? " sum-form-group--inner-cols" : "";
    return html`<div
    class=${groupClassNames(group, ctx, plain) + innerColsClass}
    style=${innerCols ? `--sum-inner-cols:${innerCols}` : false}
  >
    ${group.string ? html`<div class="sum-form-group-title">${group.string}</div>` : ""}
    <div class="sum-form-group-grid">
      ${renderFields(rf, fields, record, readonly)}
      ${renderSeparators(group.separators)}
      ${renderLabels(group.labels)}
      ${nested.map((g) => renderGroup(rf, g, record, readonly, "inner", true))}
    </div>
  </div>`;
  }
  function renderNotebook(rf, notebook, record, readonly, notebookIndex, activePage, onTab) {
    const pages = notebook.pages ?? [];
    if (pages.length === 0) return html``;
    const idx = Math.min(Math.max(activePage, 0), pages.length - 1);
    const page = pages[idx];
    return html`<div class="sum-notebook sum-notebook--sheet">
    <div class="sum-notebook-tabs" role="tablist">
      ${pages.map((pg, i) => {
      const tabClass = i === idx ? "sum-notebook-tab sum-notebook-tab--active" : "sum-notebook-tab";
      return html`<button type="button" class=${tabClass} role="tab" aria-selected=${i === idx ? "true" : "false"} @click=${() => onTab(notebookIndex, i)}>${pg.title}</button>`;
    })}
    </div>
    <div class="sum-notebook-page sum-notebook-page--sheet" role="tabpanel">
      <div class="sum-form-sheet-stack sum-notebook-page-body">
        ${renderFields(rf, page.fields ?? [], record, readonly)}
        ${(page.groups ?? []).map((g) => renderGroup(rf, g, record, readonly))}
        ${renderSeparators(page.separators)}
        ${renderLabels(page.labels)}
      </div>
    </div>
  </div>`;
  }
  function renderFormSheet(opts) {
    const {
      env,
      sheet,
      record,
      readonly,
      hasImageField = false,
      activeNotebookPages,
      onNotebookTab,
      renderField: renderFieldOpt
    } = opts;
    const rf = renderFieldOpt ?? ((f, r, ro) => renderField(env, f, r, ro));
    if (!sheet) {
      return html`<div class="sum-form-sheet"></div>`;
    }
    const parts = [];
    for (const div of sheet.divs ?? []) {
      parts.push(renderTitleDiv(rf, div, record, readonly, hasImageField));
    }
    const topFields = visibleFields(sheet.fields ?? []);
    const groups = sheet.groups ?? [];
    if (topFields.length > 0 || groups.length > 0) {
      parts.push(
        html`<div class="sum-form-sheet-stack sum-field-region--sheet">
        ${renderFields(rf, topFields, record, readonly)}
        ${groups.map((g) => renderGroup(rf, g, record, readonly))}
      </div>`
      );
    }
    (sheet.notebook ?? []).forEach((nb, notebookIndex) => {
      const activePage = activeNotebookPages[notebookIndex] ?? 0;
      parts.push(renderNotebook(rf, nb, record, readonly, notebookIndex, activePage, onNotebookTab));
    });
    const sheetSeparators = sheet.separators ?? [];
    const sheetLabels = sheet.labels ?? [];
    if (sheetSeparators.length > 0 || sheetLabels.length > 0) {
      parts.push(
        html`<div class="sum-form-sheet-meta">${renderSeparators(sheetSeparators)}${renderLabels(sheetLabels)}</div>`
      );
    }
    return html`<div class="sum-form-sheet">${parts}</div>`;
  }

  // src/views/form-interactions.ts
  function onNotebookKeydown(ev) {
    if (!(ev instanceof KeyboardEvent)) return;
    if (ev.key !== "ArrowLeft" && ev.key !== "ArrowRight") return;
    const target = ev.target;
    if (!(target instanceof HTMLButtonElement) || target.getAttribute("role") !== "tab") return;
    const tabs = target.parentElement;
    if (!tabs) return;
    const buttons = Array.from(tabs.querySelectorAll('button[role="tab"]'));
    const idx = buttons.indexOf(target);
    if (idx < 0) return;
    ev.preventDefault();
    const next = ev.key === "ArrowRight" ? Math.min(idx + 1, buttons.length - 1) : Math.max(idx - 1, 0);
    buttons[next]?.focus();
    buttons[next]?.click();
  }
  function bindMany2OneDismiss(root) {
    const onDocClick = (ev) => {
      const target = ev.target;
      if (!(target instanceof Node)) return;
      for (const widget of root.querySelectorAll(".sum-field-widget--many2one")) {
        if (widget.contains(target)) continue;
        const list = widget.querySelector(".sum-m2o-suggest");
        list?.remove();
      }
    };
    const onKey = (ev) => {
      if (ev.key !== "Escape") return;
      for (const list of root.querySelectorAll(".sum-m2o-suggest")) {
        list.remove();
      }
    };
    document.addEventListener("click", onDocClick, true);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("click", onDocClick, true);
      document.removeEventListener("keydown", onKey);
    };
  }
  function normalizeCrop(c) {
    return {
      x: Math.min(100, Math.max(0, c.x)),
      y: Math.min(100, Math.max(0, c.y)),
      zoom: Math.min(4, Math.max(1, c.zoom))
    };
  }
  function applyCropStyle(img, crop) {
    const c = normalizeCrop(crop);
    img.style.objectPosition = `${c.x}% ${c.y}%`;
    img.style.transform = `scale(${c.zoom})`;
    img.style.transformOrigin = `${c.x}% ${c.y}%`;
  }
  function openAvatarCropModal(file, onSave) {
    const modal = document.createElement("div");
    modal.className = "sum-avatar-crop-modal";
    modal.innerHTML = `
    <div class="sum-avatar-crop-modal-inner">
      <h3 class="sum-avatar-crop-title">Crop image</h3>
      <p class="sum-avatar-crop-hint">Drag to reposition \xB7 use zoom slider</p>
      <div class="sum-avatar-crop-stage">
        <div class="sum-avatar-crop-viewport">
          <img class="sum-avatar-crop-img" alt="" />
          <div class="sum-avatar-crop-ring"></div>
        </div>
      </div>
      <label class="sum-avatar-crop-zoom-label">Zoom
        <input type="range" class="sum-avatar-crop-zoom" min="1" max="4" step="0.05" value="1" />
      </label>
      <div class="sum-avatar-crop-modal-actions">
        <button type="button" class="sum-avatar-crop-save">Save</button>
        <button type="button" class="sum-avatar-crop-cancel">Cancel</button>
      </div>
    </div>`;
    const img = modal.querySelector(".sum-avatar-crop-img");
    const zoom = modal.querySelector(".sum-avatar-crop-zoom");
    const stage = modal.querySelector(".sum-avatar-crop-stage");
    let crop = { x: 50, y: 50, zoom: 1 };
    let dragging = false;
    const close = () => modal.remove();
    const reader = new FileReader();
    reader.onload = () => {
      img.src = String(reader.result ?? "");
      applyCropStyle(img, crop);
    };
    reader.readAsDataURL(file);
    stage.addEventListener("pointerdown", (ev) => {
      dragging = true;
      stage.setPointerCapture(ev.pointerId);
    });
    stage.addEventListener("pointermove", (ev) => {
      if (!dragging) return;
      const rect = stage.getBoundingClientRect();
      crop.x = (ev.clientX - rect.left) / rect.width * 100;
      crop.y = (ev.clientY - rect.top) / rect.height * 100;
      applyCropStyle(img, crop);
    });
    stage.addEventListener("pointerup", () => {
      dragging = false;
    });
    zoom.addEventListener("input", () => {
      crop.zoom = Number(zoom.value);
      applyCropStyle(img, crop);
    });
    modal.querySelector(".sum-avatar-crop-cancel")?.addEventListener("click", close);
    modal.querySelector(".sum-avatar-crop-save")?.addEventListener("click", () => {
      const canvas = document.createElement("canvas");
      const size = 256;
      canvas.width = size;
      canvas.height = size;
      const ctx = canvas.getContext("2d");
      if (ctx && img.complete) {
        const c = normalizeCrop(crop);
        const sw = img.naturalWidth / c.zoom;
        const sh = img.naturalHeight / c.zoom;
        const sx = c.x / 100 * img.naturalWidth - sw / 2;
        const sy = c.y / 100 * img.naturalHeight - sh / 2;
        ctx.drawImage(img, sx, sy, sw, sh, 0, 0, size, size);
        onSave(canvas.toDataURL("image/png"), c);
      } else {
        onSave(img.src, normalizeCrop(crop));
      }
      close();
    });
    document.body.appendChild(modal);
  }
  function bindAvatarUpload(root) {
    const onChange = (ev) => {
      const input = ev.target;
      if (!(input instanceof HTMLInputElement) || input.type !== "file") return;
      const file = input.files?.[0];
      if (!file || !file.type.startsWith("image/")) return;
      const host = input.closest("[data-sum-avatar]") ?? input.closest(".sum-field-widget--image");
      if (!host) return;
      const hidden = host.querySelector("[data-sum-avatar-value], [data-sum-image-value]");
      openAvatarCropModal(file, (dataUrl) => {
        if (hidden) hidden.value = dataUrl;
        hidden?.dispatchEvent(new Event("input", { bubbles: true }));
        const img = host.querySelector(".sum-form-avatar-img, .sum-image-thumb-img");
        if (img) {
          img.src = dataUrl;
          img.classList.add("sum-form-avatar-img--visible", "sum-form-avatar-img--cropped");
        }
        const initials2 = host.querySelector(".sum-form-avatar-initials");
        initials2?.remove();
      });
      input.value = "";
    };
    root.addEventListener("change", onChange);
    return () => root.removeEventListener("change", onChange);
  }
  function bindDateDismiss(root) {
    const onDocClick = (ev) => {
      const target = ev.target;
      if (!(target instanceof Node)) return;
      for (const details of root.querySelectorAll("details.sum-date-field[open]")) {
        if (details.contains(target)) continue;
        details.open = false;
      }
    };
    document.addEventListener("click", onDocClick, true);
    return () => document.removeEventListener("click", onDocClick, true);
  }
  function initFormInteractions(root) {
    const cleanups = [];
    for (const tabs of root.querySelectorAll(".sum-notebook-tabs")) {
      tabs.addEventListener("keydown", onNotebookKeydown);
      cleanups.push(() => tabs.removeEventListener("keydown", onNotebookKeydown));
    }
    cleanups.push(bindMany2OneDismiss(root));
    cleanups.push(bindAvatarUpload(root));
    cleanups.push(bindDateDismiss(root));
    return () => {
      for (const fn of cleanups) fn();
    };
  }

  // src/widgets/field-host.ts
  var FieldHost = class {
    env;
    entries = /* @__PURE__ */ new Map();
    constructor(env) {
      this.env = env;
    }
    render(field, record, readonly) {
      const widget = resolveFieldWidget(field);
      const key = field.name;
      const prev = this.entries.get(key);
      if (prev && prev.readonly === readonly && prev.widget === widget) {
        return prev.comp.render();
      }
      prev?.comp.destroy();
      const Ctor = registry.get("fields", widget) ?? registry.get("fields", "default");
      const comp = new Ctor({ field, record, readonly }, this.env);
      comp.setup?.();
      this.entries.set(key, { comp, readonly, widget });
      return comp.render();
    }
    clear() {
      for (const { comp } of this.entries.values()) {
        comp.destroy();
      }
      this.entries.clear();
    }
  };

  // src/views/FormView.ts
  var FormView = class extends SwcComponent {
    recordStore;
    record;
    snapshot = {};
    editing = false;
    saving = false;
    acting = false;
    error = "";
    activeNotebookPages = {};
    teardownInteractions = null;
    fieldHost;
    setup() {
      this.bump = () => {
        if (this.el?.isConnected) this.patch();
      };
      this.recordStore = new RecordStore(this.env.services.rpc);
      this.fieldHost = new FieldHost(this.env);
      const p = this.props.payload;
      this.editing = p.formEdit || p.recordId <= 0;
      this.snapshot = { ...p.record ?? {} };
      this.record = this.recordStore.fromPayload(p.model, p.recordId, this.snapshot);
    }
    bump = null;
    onMount() {
      this.bindFormInteractions();
    }
    onWillUnmount() {
      this.teardownInteractions?.();
      this.teardownInteractions = null;
      this.fieldHost.clear();
    }
    patch() {
      this.teardownInteractions?.();
      if (!this.el?.parentElement) return;
      const parent = this.el.parentElement;
      const next = this.template().render();
      parent.replaceChild(next, this.el);
      this.el = next;
      this.bindFormInteractions();
    }
    bindFormInteractions() {
      if (this.el) {
        this.teardownInteractions = initFormInteractions(this.el);
      }
    }
    renderFieldCached = (field, record, readonly) => this.fieldHost.render(field, record, readonly);
    isReadonly() {
      return !this.editing;
    }
    toolbarBusy() {
      return this.saving || this.acting;
    }
    fields() {
      const arch = this.props.payload.arch;
      return collectFormFields(arch.sheet, arch.header?.fields ?? []);
    }
    headerButtons() {
      return this.props.payload.arch.header?.buttons ?? [];
    }
    startEdit() {
      this.editing = true;
      this.error = "";
      this.bump?.();
    }
    cancelEdit() {
      const p = this.props.payload;
      if (p.recordId <= 0) {
        const url = this.env.services.router.workspaceUrl({
          actionId: p.actionId,
          menuId: p.menuId,
          viewType: "list",
          recordId: 0,
          formEdit: false
        });
        this.env.services.action.navigate(url);
        return;
      }
      this.record = this.recordStore.fromPayload(p.model, p.recordId, { ...this.snapshot });
      this.editing = false;
      this.error = "";
      this.bump?.();
    }
    async reloadRecord() {
      const p = this.props.payload;
      if (p.recordId <= 0) return;
      const fieldNames = this.fields().map((f) => f.name);
      const rows = await this.env.services.rpc.read(p.model, [p.recordId], fieldNames);
      if (!rows[0]) return;
      this.snapshot = { ...rows[0] };
      this.record = this.recordStore.fromPayload(p.model, p.recordId, this.snapshot);
      this.bump?.();
    }
    async save() {
      this.saving = true;
      this.error = "";
      this.bump?.();
      try {
        const required = this.fields().filter((f) => f.required).map((f) => f.name);
        this.recordStore.validate(this.record, required);
        const id = await this.recordStore.save(this.record);
        this.env.services.notification.show({
          kind: "success",
          title: "Saved",
          body: "Record saved successfully."
        });
        const p = this.props.payload;
        if (p.recordId <= 0 && id > 0) {
          this.env.services.action.openRecord(p.model, p.actionId, p.menuId, id, "form");
          return;
        }
        this.snapshot = { ...this.record.data };
        this.editing = false;
        this.bump?.();
      } catch (err) {
        this.error = err instanceof SwcError ? err.message : String(err);
      } finally {
        this.saving = false;
        this.bump?.();
      }
    }
    async runObjectButton(btn) {
      const p = this.props.payload;
      if (btn.type !== "object" || p.recordId <= 0) return;
      this.acting = true;
      this.error = "";
      this.bump?.();
      try {
        const result = await this.env.services.rpc.callMethod(p.model, btn.name, p.recordId);
        if (result?.redirect) {
          this.env.services.action.navigate(result.redirect);
          return;
        }
        this.env.services.notification.show({
          kind: "success",
          title: btn.string || btn.name,
          body: "Action completed."
        });
        await this.reloadRecord();
      } catch (err) {
        this.error = err instanceof SwcError ? err.message : String(err);
      } finally {
        this.acting = false;
        this.bump?.();
      }
    }
    renderToolbarPrimary() {
      const p = this.props.payload;
      const busy = this.toolbarBusy();
      const items = [];
      if (p.recordId > 0 && this.isReadonly()) {
        items.push(renderNewButton(p));
        items.push(headerButton("Edit", void 0, () => this.startEdit(), busy));
      } else {
        items.push(
          headerButton("Save", "sum_highlight", () => void this.save(), busy)
        );
        items.push(headerButton("Cancel", void 0, () => this.cancelEdit(), busy || this.saving));
      }
      for (const btn of this.headerButtons()) {
        if (btn.type !== "object") continue;
        items.push(
          headerButton(btn.string || btn.name, btn.class, () => void this.runObjectButton(btn), busy)
        );
      }
      return items;
    }
    template() {
      const p = this.props.payload;
      const readonly = this.isReadonly();
      const headerFields = p.arch.header?.fields ?? [];
      const exportFields = visibleFieldNames(this.fields());
      const reportActions = p.recordId > 0 ? renderReportActions(p, exportFields, p.recordId) : null;
      const toolbarItems = this.renderToolbarPrimary();
      const busy = this.toolbarBusy();
      const sheet = renderFormSheet({
        env: this.env,
        sheet: p.arch.sheet,
        record: this.record,
        readonly,
        hasImageField: p.arch.formMeta?.hasImageField ?? false,
        activeNotebookPages: this.activeNotebookPages,
        onNotebookTab: (notebookIndex, pageIndex) => {
          this.activeNotebookPages = { ...this.activeNotebookPages, [notebookIndex]: pageIndex };
          this.bump?.();
        },
        renderField: this.renderFieldCached
      });
      const footerButtons = p.arch.footer?.buttons ?? [];
      return html`
      <div class="sum-form-view sum-form-view--workspace-chrome${readonly ? " sum-form-view--readonly" : ""}">
        <div class="sum-ws-record-toolbar sum-view-toolbar sum-form-toolbar">
          <div class="sum-statusbar-buttons sum-view-toolbar-primary">${toolbarItems}</div>
          ${headerFields.length > 0 ? html`<div class="sum-statusbar-status sum-ws-toolbar-right">
                ${headerFields.map((f) => this.renderFieldCached(f, this.record, readonly))}
              </div>` : ""}
          ${reportActions ?? ""}
        </div>
        ${this.error ? html`<div class="sum-flash sum-flash--error">${this.error}</div>` : ""}
        <div class="sum-form-sheet-bg">
          ${sheet}
          ${footerButtons.length > 0 ? html`<div class="sum-form-footer">
                ${footerButtons.map(
        (btn) => headerButton(
          btn.string || btn.name,
          btn.class,
          () => void this.runObjectButton(btn),
          busy
        )
      )}
              </div>` : ""}
        </div>
      </div>
    `;
    }
  };

  // src/views/kanban-card.ts
  function isKanbanImageField(field) {
    const name = field.name.toLowerCase();
    return name === "image" || name.startsWith("image_") || field.widget === "image" || field.widget === "circle";
  }
  function isKanbanImageCircle(field) {
    return field.widget === "circle" || field.options?.shape === "circle";
  }
  function isPriorityField(field) {
    return field.name === "priority" || field.widget === "priority";
  }
  function displayValue(row, field) {
    const raw = row[`${field.name}_name`] ?? row[field.name];
    if (raw == null || raw === false) return "";
    return String(raw);
  }
  function imageSrc(row, field) {
    const raw = row[field.name];
    if (typeof raw !== "string" || !raw.trim()) return "";
    const v = raw.trim();
    if (v.startsWith("data:") || v.startsWith("http://") || v.startsWith("https://") || v.startsWith("/")) {
      return v;
    }
    return "";
  }
  function initials(row, fields) {
    const nameField = fields.find((f) => f.name === "name") ?? fields.find((f) => !isKanbanImageField(f));
    const text = nameField ? displayValue(row, nameField) : "";
    const parts = text.trim().split(/\s+/).filter(Boolean);
    if (parts.length === 0) return "?";
    if (parts.length === 1) return parts[0].slice(0, 1).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }
  function titleField(fields) {
    return fields.find((f) => f.name === "name") ?? fields.find((f) => f.name === "display_name") ?? fields.find((f) => !isKanbanImageField(f) && !isPriorityField(f));
  }
  function renderPriority(row, field) {
    const level = Number(row[field.name] ?? 0);
    if (!level) return null;
    const stars = [1, 2, 3].map(
      (n) => html`<span class="sum-kanban-priority-star${n <= level ? " sum-kanban-priority-star--on" : ""}">★</span>`
    );
    return html`<div class="sum-kanban-priority">${stars}</div>`;
  }
  function renderMedia(row, imageField, fields) {
    const src = imageSrc(row, imageField);
    const label = displayValue(row, titleField(fields) ?? imageField);
    if (!src && !label) return null;
    const media = document.createElement("div");
    media.className = `sum-kanban-card-media${isKanbanImageCircle(imageField) ? " sum-kanban-card-media--circle" : " sum-kanban-card-media--square"}`;
    if (src) {
      const img = document.createElement("img");
      img.className = "sum-kanban-card-media-img";
      img.src = src;
      img.alt = "";
      media.appendChild(img);
    } else {
      const initialsEl = document.createElement("span");
      initialsEl.className = "sum-kanban-card-media-initials";
      initialsEl.textContent = initials(row, fields);
      media.appendChild(initialsEl);
    }
    return media;
  }
  function renderKanbanCardInner(row, fields) {
    const imageField = fields.find(isKanbanImageField);
    const priorityField = fields.find(isPriorityField);
    const title = titleField(fields);
    const subs = fields.filter(
      (f) => f !== imageField && f !== title && f !== priorityField && !isKanbanImageField(f) && !isPriorityField(f)
    );
    const media = imageField ? renderMedia(row, imageField, fields) : null;
    const titleEl = title ? html`<div class="sum-kanban-card-title">${displayValue(row, title)}</div>` : null;
    const subEls = subs.map((f) => displayValue(row, f)).filter(Boolean).map((text) => html`<div class="sum-kanban-card-sub">${text}</div>`);
    const priorityEl = priorityField ? renderPriority(row, priorityField) : null;
    if (media) {
      return html`${media}<div class="sum-kanban-card-body">${titleEl}${subEls}${priorityEl}</div>`;
    }
    return html`${titleEl}${subEls}${priorityEl}`;
  }

  // src/views/KanbanView.ts
  var KanbanView = class extends SwcComponent {
    cardFields() {
      return this.props.payload.arch.fields.filter((f) => !f.invisible);
    }
    openCard(row) {
      const id = Number(row.id ?? 0);
      if (id <= 0) return;
      const p = this.props.payload;
      this.env.services.action.openRecord(p.model, p.actionId, p.menuId, id, "form");
    }
    async moveCard(recordId, columnValue) {
      const groupField = this.props.payload.arch.kanban?.groupField;
      if (!groupField) return;
      await this.env.services.rpc.write(this.props.payload.model, [recordId], {
        [groupField]: columnValue || false
      });
      window.location.reload();
    }
    toolbar() {
      const p = this.props.payload;
      const fields = visibleFieldNames(this.cardFields());
      const reportActions = renderReportActions(p, fields);
      return html`
      <div class="sum-view-toolbar sum-kanban-report-bar">
        <div class="sum-view-toolbar-primary">${renderNewButton(p)}</div>
        ${reportActions ?? ""}
      </div>
    `;
    }
    renderCard(row, fields, draggable = false) {
      return html`<div
      class="sum-kanban-card"
      draggable=${draggable ? "true" : void 0}
      @click=${() => this.openCard(row)}
      @dragstart=${draggable ? (ev) => ev.dataTransfer?.setData("text/plain", String(row.id)) : void 0}
    >
      ${renderKanbanCardInner(row, fields)}
    </div>`;
    }
    template() {
      const p = this.props.payload;
      const kanban = p.arch.kanban;
      const fields = this.cardFields();
      if (!kanban?.columns?.length) {
        const rows = p.records ?? [];
        return html`
        <div class="sum-kanban-view">
          ${this.toolbar()}
          <div class="sum-kanban-columns">
            ${rows.length === 0 ? html`<div class="sum-kanban-empty">No records</div>` : rows.map((row) => this.renderCard(row, fields))}
          </div>
        </div>
      `;
      }
      return html`
      <div class="sum-kanban-view">
        ${this.toolbar()}
        <div class="sum-kanban-board sum-kanban-board--grouped">
          <div class="sum-kanban-stage-columns">
            ${kanban.columns.map(
        (col) => html`<div class="sum-kanban-stage-column" data-column=${String(col.value)}>
                <div class="sum-kanban-stage-header">
                  <span>${col.label}</span>
                  <span class="sum-kanban-stage-count">${String(col.records.length)}</span>
                </div>
                <div class="sum-kanban-cards">
                  ${col.records.map(
          (row) => html`<div
                      class="sum-kanban-card"
                      draggable=${kanban.draggable ? "true" : void 0}
                      @click=${() => this.openCard(row)}
                      @dragstart=${(ev) => ev.dataTransfer?.setData("text/plain", String(row.id))}
                      @dragover=${(ev) => ev.preventDefault()}
                      @drop=${(ev) => {
            ev.preventDefault();
            const de = ev;
            const id = Number(de.dataTransfer?.getData("text/plain"));
            if (id) void this.moveCard(id, col.value);
          }}
                    >
                      ${renderKanbanCardInner(row, fields)}
                    </div>`
        )}
                </div>
              </div>`
      )}
          </div>
        </div>
      </div>
    `;
    }
  };

  // src/views/PivotView.ts
  var PivotView = class extends SwcComponent {
    template() {
      const pivot = this.props.payload.arch.pivot;
      if (!pivot) {
        return html`<div class="sum-pivot-view sum-pivot-view--empty">No pivot data</div>`;
      }
      return html`
      <div class="sum-pivot-view">
        <table class="sum-pivot-table">
          <thead>
            <tr>
              <th></th>
              ${pivot.colLabels.map((c) => html`<th>${c}</th>`)}
            </tr>
          </thead>
          <tbody>
            ${pivot.rowLabels.map(
        (row) => html`<tr>
                <th>${row}</th>
                ${pivot.colLabels.map((col) => {
          const val = pivot.values[row]?.[col] ?? 0;
          return html`<td>${String(val)}</td>`;
        })}
              </tr>`
      )}
          </tbody>
        </table>
        <p class="sum-pivot-measure">${pivot.measureLabel}</p>
      </div>
    `;
    }
  };

  // src/views/GraphView.ts
  var GraphView = class extends SwcComponent {
    groups = [];
    measureField = "id";
    setup() {
      const [, bump] = useState(0);
      this.bump = () => bump((n) => n + 1);
      useEffect(() => {
        void this.load();
      });
    }
    bump = null;
    async load() {
      const p = this.props.payload;
      const groupField = p.arch.fields.find((f) => f.pivotType === "row")?.name ?? "create_date";
      this.measureField = p.arch.fields.find((f) => f.pivotType === "measure")?.name ?? "id";
      this.groups = await this.env.services.rpc.readGroup(p.model, [], [this.measureField], [groupField], 40);
      this.bump?.();
    }
    template() {
      const max = Math.max(...this.groups.map((g) => Number(g[this.measureField] ?? 0)), 1);
      return html`
      <div class="sum-graph-view">
        ${this.groups.map((g) => {
        const label = String(g[`${Object.keys(g).find((k) => k.endsWith("_count")) ?? "name"}`] ?? g.name ?? "");
        const val = Number(g[this.measureField] ?? 0);
        const pct = Math.round(val / max * 100);
        return html`<div class="sum-graph-bar-row">
            <span class="sum-graph-label">${label}</span>
            <div class="sum-graph-bar" style="width:${pct}%"></div>
            <span class="sum-graph-value">${val}</span>
          </div>`;
      })}
      </div>
    `;
    }
  };

  // src/views/WorkspaceRouter.ts
  var WorkspaceRouter = class extends SwcComponent {
    payload = null;
    loading = true;
    error = "";
    setup() {
      const [, bump] = useState(0);
      this.bump = () => bump((n) => n + 1);
      const load = async () => {
        this.loading = true;
        this.error = "";
        this.bump?.();
        try {
          this.payload = await this.fetchWorkspace();
        } catch (err) {
          this.error = err instanceof SwcError ? err.message : String(err);
        } finally {
          this.loading = false;
          this.bump?.();
        }
      };
      void load();
      useEffect(() => {
        const onNav = () => void load();
        window.addEventListener("popstate", onNav);
        return () => window.removeEventListener("popstate", onNav);
      });
    }
    bump = null;
    async fetchWorkspace() {
      const route = this.env.services.router.parse();
      const params = new URLSearchParams();
      if (route.actionId) params.set("action", String(route.actionId));
      if (route.menuId) params.set("menu_id", route.menuId);
      if (route.viewType) params.set("view_type", route.viewType);
      if (route.recordId) params.set("id", String(route.recordId));
      if (route.formEdit) params.set("edit", "1");
      if (route.listSearch) params.set("q", route.listSearch);
      const base = this.env.bootstrap.swcApiBase || "/web/swc";
      return this.env.services.http.getJSON(`${base}/workspace?${params.toString()}`);
    }
    mountView(view) {
      view.setup?.();
      return view.render();
    }
    renderView() {
      if (!this.payload) return document.createElement("div");
      const type = this.payload.viewType || this.payload.arch.type;
      const p = this.payload;
      switch (type) {
        case "form":
          return this.mountView(new FormView({ payload: p }, this.env));
        case "kanban":
          return this.mountView(new KanbanView({ payload: p }, this.env));
        case "pivot":
          return this.mountView(new PivotView({ payload: p }, this.env));
        case "graph":
          return this.mountView(new GraphView({ payload: p }, this.env));
        default:
          return this.mountView(new ListView({ payload: p }, this.env));
      }
    }
    template() {
      if (this.loading) {
        return html`<div class="sum-workspace-loading">Loading workspace…</div>`;
      }
      if (this.error) {
        return html`<div class="sum-flash sum-flash--error">${this.error}</div>`;
      }
      return html`<div class="sum-workspace-view">${this.renderView()}</div>`;
    }
  };

  // src/shell/ShellLayout.ts
  var ShellLayout = class extends SwcComponent {
    launcherOpen = false;
    workspaceRouter;
    appLauncher;
    setup() {
      const [, setTick] = useState(0);
      const bump = () => setTick((n) => n + 1);
      this.workspaceRouter = new WorkspaceRouter({}, this.env);
      this.workspaceRouter.setup?.();
      const boot = this.env.bootstrap;
      this.appLauncher = new AppLauncher(
        {
          apps: boot.apps,
          isOpen: () => this.launcherOpen,
          requestClose: () => {
            this.launcherOpen = false;
            bump();
          }
        },
        this.env
      );
      this.appLauncher.setup?.();
      useEffect(() => {
        const toggleLauncher = () => {
          this.launcherOpen = !this.launcherOpen;
          bump();
        };
        const onKey = (ev) => {
          if ((ev.ctrlKey || ev.metaKey) && ev.key.toLowerCase() === "k") {
            ev.preventDefault();
            toggleLauncher();
          }
        };
        document.addEventListener("keydown", onKey);
        document.addEventListener("swc:launcher-toggle", toggleLauncher);
        const searchBtn = document.getElementById("sum-topbar-search-open");
        searchBtn?.addEventListener("click", toggleLauncher);
        return () => {
          document.removeEventListener("keydown", onKey);
          document.removeEventListener("swc:launcher-toggle", toggleLauncher);
          searchBtn?.removeEventListener("click", toggleLauncher);
        };
      });
      if (this.env.bootstrap.busEnabled) {
        this.env.services.bus.connect();
      }
    }
    workspaceView() {
      if (this.workspaceRouter.el?.isConnected) {
        this.workspaceRouter.patch();
        return this.workspaceRouter.el;
      }
      return this.workspaceRouter.render();
    }
    template() {
      return html`
      <div id="swc-root-inner">
        <main class="sum-workspace-inner">
          ${this.workspaceView()}
        </main>
        ${this.appLauncher.render()}
      </div>
    `;
    }
  };

  // src/util/shell-storage.ts
  var KEY_SIDEBAR = "sum.shell.sidebarCollapsed";
  var KEY_ACTIVITY_WIDTH = "sum.shell.activityWidthPx";
  var KEY_ACTIVITY_HIDDEN = "sum.shell.activityHidden";
  function readBool(key) {
    try {
      return localStorage.getItem(key) === "1";
    } catch {
      return false;
    }
  }
  function writeBool(key, value) {
    try {
      localStorage.setItem(key, value ? "1" : "0");
    } catch {
    }
  }
  function readActivityWidth() {
    try {
      const n = parseInt(localStorage.getItem(KEY_ACTIVITY_WIDTH) ?? "", 10);
      if (n >= 200 && n <= 520) return n;
    } catch {
    }
    return 300;
  }
  function writeActivityWidth(px) {
    try {
      localStorage.setItem(KEY_ACTIVITY_WIDTH, String(Math.round(px)));
    } catch {
    }
  }
  function readJSON(key, fallback) {
    try {
      const raw = localStorage.getItem(key);
      if (!raw) return fallback;
      const value = JSON.parse(raw);
      return Array.isArray(value) ? value : fallback;
    } catch {
      return fallback;
    }
  }

  // src/shell/activity-panel.ts
  var CHEVRON_LEFT = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M15 18l-6-6 6-6"/></svg>';
  function applyActivityWidth(px) {
    document.documentElement.style.setProperty("--sum-activity-width", `${px}px`);
  }
  function applyActivityHidden(shell, hidden) {
    shell.classList.toggle("sum-shell--activity-hidden", hidden);
    const reveal = document.getElementById("sum-activity-reveal");
    if (reveal) reveal.hidden = !hidden;
    const toggle = document.getElementById("sum-activity-toggle");
    if (toggle) {
      const pressed = !hidden;
      toggle.setAttribute("aria-pressed", pressed ? "true" : "false");
      toggle.classList.toggle("is-pressed", pressed);
    }
  }
  function paintActivityRevealIcon() {
    const reveal = document.getElementById("sum-activity-reveal");
    if (reveal && !reveal.firstChild) reveal.innerHTML = CHEVRON_LEFT;
  }
  function initActivityTabs() {
    document.querySelectorAll("[data-sum-activity-tab]").forEach((tab) => {
      tab.addEventListener("click", () => {
        const name = tab.getAttribute("data-sum-activity-tab");
        const panes = {
          messages: "sum-activity-pane-messages",
          log: "sum-activity-pane-log"
        };
        document.querySelectorAll("[data-sum-activity-tab]").forEach((t) => {
          const on = t.getAttribute("data-sum-activity-tab") === name;
          t.classList.toggle("is-active", on);
          t.setAttribute("aria-selected", on ? "true" : "false");
        });
        for (const [key, id] of Object.entries(panes)) {
          const el = document.getElementById(id);
          if (el) el.hidden = key !== name;
        }
      });
    });
  }
  function initActivityResizer(shell) {
    const resizer = document.getElementById("sum-activity-resizer");
    if (!resizer) return;
    let dragging = false;
    let startX = 0;
    let startW = 300;
    resizer.addEventListener("mousedown", (ev) => {
      if (shell.classList.contains("sum-shell--activity-hidden")) return;
      dragging = true;
      startX = ev.clientX;
      startW = readActivityWidth();
      document.body.style.cursor = "col-resize";
      document.body.style.userSelect = "none";
      ev.preventDefault();
    });
    window.addEventListener("mousemove", (ev) => {
      if (!dragging) return;
      const delta = startX - ev.clientX;
      let width = startW + delta;
      width = Math.min(520, Math.max(200, width));
      applyActivityWidth(width);
    });
    window.addEventListener("mouseup", () => {
      if (!dragging) return;
      dragging = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      const raw = document.documentElement.style.getPropertyValue("--sum-activity-width");
      const px = parseInt(raw, 10);
      if (!Number.isNaN(px)) writeActivityWidth(px);
    });
  }
  function initActivityPanel(shell) {
    applyActivityWidth(readActivityWidth());
    applyActivityHidden(shell, readBool(KEY_ACTIVITY_HIDDEN));
    paintActivityRevealIcon();
    const setActivityHidden = (hidden) => {
      applyActivityHidden(shell, hidden);
      writeBool(KEY_ACTIVITY_HIDDEN, hidden);
    };
    document.getElementById("sum-activity-toggle")?.addEventListener("click", () => {
      setActivityHidden(!shell.classList.contains("sum-shell--activity-hidden"));
    });
    document.getElementById("sum-activity-collapse")?.addEventListener("click", () => {
      setActivityHidden(true);
    });
    document.getElementById("sum-activity-reveal")?.addEventListener("click", () => {
      setActivityHidden(false);
    });
    initActivityTabs();
    initActivityResizer(shell);
  }

  // src/shell/pinned-apps.ts
  var KEY_PINNED_LEGACY = "sumeru:pinned-apps";
  var pinnedCache = [];
  var cacheLoaded = false;
  async function persistPinnedApps(http, modules) {
    const data = await http.postJSON("/web/user/pinned-apps", { modules });
    return Array.isArray(data.modules) ? data.modules.map(String) : modules;
  }
  function loadPinnedApps(initial) {
    if (!cacheLoaded) {
      pinnedCache = initial.slice();
      cacheLoaded = true;
    }
    return pinnedCache.slice();
  }
  function setPinnedCache(modules) {
    pinnedCache = modules.slice();
    cacheLoaded = true;
  }
  function getPinnedApps() {
    return pinnedCache.slice();
  }
  function togglePinnedApp(http, moduleName) {
    const mod = String(moduleName || "").trim();
    if (!mod) return getPinnedApps();
    const previous = getPinnedApps();
    let pins = previous.slice();
    if (pins.includes(mod)) {
      pins = pins.filter((m) => m !== mod);
    } else {
      pins = [mod, ...pins];
    }
    setPinnedCache(pins);
    persistPinnedApps(http, pins).then((saved) => {
      setPinnedCache(saved);
    }).catch(() => {
      setPinnedCache(previous);
    });
    return pins;
  }
  function applyTopNavFilter() {
    const nav = document.querySelector(".sum-top-nav");
    if (!nav) return;
    const moduleItems = [...nav.querySelectorAll(".top-menu-item--module")];
    if (moduleItems.length === 0) return;
    const pins = getPinnedApps();
    const visibleMods = new Set(pins);
    moduleItems.forEach((el) => {
      const mod = el.getAttribute("data-module") ?? "";
      el.classList.toggle("is-topbar-hidden", !visibleMods.has(mod));
    });
    const activeEl = nav.querySelector(".top-menu-item.active");
    activeEl?.scrollIntoView?.({ inline: "nearest", block: "nearest", behavior: "instant" });
  }
  function initPinnedApps(http, initial) {
    loadPinnedApps(initial);
    const legacy = readJSON(KEY_PINNED_LEGACY, []);
    if (getPinnedApps().length === 0 && legacy.length > 0) {
      persistPinnedApps(http, legacy).then((saved) => {
        setPinnedCache(saved);
        try {
          localStorage.removeItem(KEY_PINNED_LEGACY);
        } catch {
        }
        applyTopNavFilter();
      }).catch(() => {
      });
    }
    applyTopNavFilter();
  }

  // src/shell/home-dashboard.ts
  function updatePinButton(btn, displayName, pinned) {
    const name = displayName || btn.getAttribute("data-module") || "App";
    btn.classList.toggle("is-pinned", pinned);
    btn.setAttribute("aria-pressed", pinned ? "true" : "false");
    const label = pinned ? `Unpin ${name} from top bar` : `Pin ${name} to top bar`;
    btn.setAttribute("aria-label", label);
    btn.setAttribute("title", pinned ? "Pinned to top bar \u2014 click to unpin" : "Pin to top bar");
  }
  function syncAllPinButtons() {
    const pins = getPinnedApps();
    document.querySelectorAll(".sum-home-hub-app-pin").forEach((btn) => {
      const mod = btn.getAttribute("data-module") ?? "";
      updatePinButton(btn, btn.getAttribute("data-display-name") ?? mod, pins.includes(mod));
    });
  }
  function tileDisplayName(tile) {
    const nameEl = tile.querySelector(".sum-home-hub-app-name");
    return (nameEl?.textContent ?? tile.getAttribute("data-module") ?? "").trim().toLowerCase();
  }
  function sortTilesAZ(tiles) {
    return [...tiles].sort((a, b) => {
      const displayA = tileDisplayName(a);
      const displayB = tileDisplayName(b);
      if (displayA !== displayB) return displayA.localeCompare(displayB);
      return (a.getAttribute("data-module") ?? "").localeCompare(b.getAttribute("data-module") ?? "");
    });
  }
  function organizePinnedGrid() {
    const pinnedSection = document.getElementById("sum-home-pinned-section");
    const pinnedContainer = document.getElementById("sum-home-pinned-apps");
    const allContainer = document.getElementById("sum-home-all-apps");
    if (!pinnedSection || !pinnedContainer || !allContainer) return;
    const pins = getPinnedApps();
    const allTiles = [
      ...pinnedContainer.querySelectorAll(".sum-home-hub-app"),
      ...allContainer.querySelectorAll(".sum-home-hub-app")
    ];
    const tilesByModule = {};
    allTiles.forEach((tile) => {
      const mod = tile.getAttribute("data-module");
      if (mod) tilesByModule[mod] = tile;
    });
    const pinnedTiles = pins.map((mod) => tilesByModule[mod]).filter(Boolean);
    sortTilesAZ(pinnedTiles).forEach((tile) => {
      pinnedContainer.appendChild(tile);
    });
    sortTilesAZ(allTiles.filter((tile) => !pins.includes(tile.getAttribute("data-module") ?? ""))).forEach(
      (tile) => {
        allContainer.appendChild(tile);
      }
    );
    pinnedSection.hidden = pinnedContainer.children.length === 0;
  }
  function showHomeToast(message) {
    const toast = document.getElementById("sum-home-toast");
    if (!toast) return;
    toast.textContent = message;
    toast.hidden = false;
    window.setTimeout(() => {
      toast.hidden = true;
    }, 3200);
  }
  function initHomeDashboard(http) {
    if (!document.getElementById("sum-home-hub")) return;
    document.addEventListener(
      "click",
      (ev) => {
        const btn = ev.target?.closest(".sum-home-hub-app-pin");
        if (!btn) return;
        ev.preventDefault();
        ev.stopPropagation();
        const mod = btn.getAttribute("data-module") ?? "";
        const displayName = btn.getAttribute("data-display-name") || mod;
        const pins = togglePinnedApp(http, mod);
        const pinned = pins.includes(mod);
        updatePinButton(btn, displayName, pinned);
        syncAllPinButtons();
        organizePinnedGrid();
        applyTopNavFilter();
        showHomeToast(
          pinned ? `${displayName} pinned to top bar` : `${displayName} unpinned from top bar`
        );
      },
      true
    );
    syncAllPinButtons();
    organizePinnedGrid();
  }

  // src/shell/sidebar.ts
  var CHEVRON_RIGHT = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M9 18l6-6-6-6"/></svg>';
  function applySidebar(shell, collapsed) {
    shell.classList.toggle("sum-shell--sidebar-collapsed", collapsed);
    const reveal = document.getElementById("sum-sidebar-reveal");
    if (reveal) reveal.hidden = !collapsed;
  }
  function paintSidebarRevealIcon() {
    const reveal = document.getElementById("sum-sidebar-reveal");
    if (reveal && !reveal.firstChild) reveal.innerHTML = CHEVRON_RIGHT;
  }
  function initSidebar(shell) {
    applySidebar(shell, readBool(KEY_SIDEBAR));
    paintSidebarRevealIcon();
    const toggleSidebar = () => {
      const next = !shell.classList.contains("sum-shell--sidebar-collapsed");
      applySidebar(shell, next);
      writeBool(KEY_SIDEBAR, next);
    };
    for (const id of ["sum-sidebar-toggle", "sum-sidebar-toggle-breadcrumb"]) {
      document.getElementById(id)?.addEventListener("click", toggleSidebar);
    }
    document.getElementById("sum-sidebar-reveal")?.addEventListener("click", () => {
      applySidebar(shell, false);
      writeBool(KEY_SIDEBAR, false);
    });
  }

  // src/shell/shell-chrome.ts
  function initShellChrome(boot, http) {
    const shell = document.getElementById("sum-shell");
    if (!shell) return;
    initSidebar(shell);
    if (boot.activityEnabled) {
      initActivityPanel(shell);
    }
    initPinnedApps(http, boot.pinnedApps ?? []);
    initHomeDashboard(http);
    new NotificationService().bootstrap(boot.toasts);
  }

  // src/addon/loader.ts
  var AddonLoader = class _AddonLoader {
    static async loadEntries(urls) {
      for (const url of urls) {
        try {
          await import(
            /* @vite-ignore */
            url
          );
        } catch (err) {
          console.warn("SWC addon entry failed:", url, err);
        }
      }
    }
    static registerFromGlobal() {
      const entries = window.__SWC_ADDON_ENTRIES__;
      if (entries?.length) {
        void _AddonLoader.loadEntries(entries);
      }
    }
  };

  // src/main.ts
  function registerCore() {
    registerDefaultWidgets();
    const views = registry.category("views");
    views.add("list", ListView);
    views.add("form", FormView);
    views.add("kanban", KanbanView);
    views.add("pivot", PivotView);
    views.add("graph", GraphView);
    const main = registry.category("main_components");
    main.add("shell", ShellLayout);
  }
  function buildEnv(boot) {
    const services = {
      rpc: new RpcService(boot.rpcUrl, boot.csrfToken),
      http: new HttpService(boot.csrfToken),
      notification: new NotificationService(),
      action: new ActionService(),
      router: new RouterService(),
      bus: new BusService()
    };
    return new SwcEnv(boot, services);
  }
  function bootstrap() {
    registerCore();
    AddonLoader.registerFromGlobal();
    let boot;
    try {
      boot = readBootstrap();
    } catch {
      return;
    }
    const env = buildEnv(boot);
    initShellChrome(boot, env.services.http);
    const mountEl = document.getElementById("swc-workspace");
    if (mountEl) {
      SwcApp.start(mountEl, env, ShellLayout);
    }
  }
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", bootstrap);
  } else {
    bootstrap();
  }
  return __toCommonJS(main_exports);
})();
//# sourceMappingURL=swc.js.map
