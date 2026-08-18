type EffectCleanup = () => void;
type EffectFn = () => void | EffectCleanup;

const mountCallbacks: EffectFn[] = [];
const unmountCallbacks: EffectFn[] = [];
let activeComponent: { schedulePatch: () => void } | null = null;

export function setActiveComponent(comp: { schedulePatch: () => void } | null): void {
  activeComponent = comp;
}

export function runMountCallbacks(): void {
  for (const fn of mountCallbacks.splice(0)) {
    fn();
  }
}

export function runUnmountCallbacks(): void {
  for (const fn of unmountCallbacks.splice(0)) {
    fn();
  }
}

export function onMount(fn: EffectFn): void {
  mountCallbacks.push(fn);
}

export function onWillUnmount(fn: EffectFn): void {
  unmountCallbacks.push(fn);
}

export function useState<T>(initial: T): [() => T, (v: T | ((prev: T) => T)) => void] {
  let value = initial;
  const set = (next: T | ((prev: T) => T)): void => {
    value = typeof next === "function" ? (next as (prev: T) => T)(value) : next;
    activeComponent?.schedulePatch();
  };
  return [() => value, set];
}

export function useRef<T>(initial: T): { current: T } {
  const ref = { current: initial };
  return ref;
}

export function useEffect(fn: EffectFn): void {
  onMount(() => {
    const cleanup = fn();
    if (typeof cleanup === "function") {
      onWillUnmount(cleanup);
    }
  });
}
