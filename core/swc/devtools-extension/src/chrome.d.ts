/** Minimal Chrome extension API surface for SWC Vision. */

declare namespace chrome {
  namespace devtools {
    namespace panels {
      function create(
        title: string,
        iconPath: string,
        pagePath: string,
        callback?: () => void,
      ): void;
    }
    namespace inspectedWindow {
      function eval<T>(
        expression: string,
        callback?: (result: T, exceptionInfo: unknown) => void,
      ): void;
    }
  }
}
