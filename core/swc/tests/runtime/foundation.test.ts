import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { DialogService } from "../../src/services/dialog.js";
import { ActionService } from "../../src/services/action.js";
import { RouterService } from "../../src/services/router.js";
import { registry } from "../../src/runtime/registry.js";
import { registerCoreServices } from "../../src/services/service-registry.js";
import { BusService } from "../../src/services/bus.js";

describe("DialogService", () => {
  afterEach(() => {
    document.querySelector(".sum-dialog-layer")?.remove();
  });

  it("opens confirm dialog and resolves true", async () => {
    const dialog = new DialogService();
    const promise = dialog.confirm("Title", "Body");
    const layer = document.querySelector(".sum-dialog-layer");
    expect(layer).toBeTruthy();
    const buttons = layer!.querySelectorAll("button");
    buttons[buttons.length - 1]?.click();
    await expect(promise).resolves.toBe(true);
  });
});

describe("ActionService SPA navigation", () => {
  it("uses router.push for /web? URLs", () => {
    const router = new RouterService();
    const push = vi.spyOn(router, "push");
    const action = new ActionService(router);
    action.navigate("/web?action=1&menu_id=m&view_type=list");
    expect(push).toHaveBeenCalled();
  });
});

describe("service registry", () => {
  it("registers core services", () => {
    const services = {
      rpc: {} as never,
      http: {} as never,
      notification: {} as never,
      action: {} as never,
      router: {} as never,
      bus: new BusService(),
      dialog: new DialogService(),
    };
    registerCoreServices(services);
    expect(registry.get("services", "dialog")).toBe(services.dialog);
  });
});

describe("BusService record.updated", () => {
  it("notifies subscribers", () => {
    const bus = new BusService();
    const fn = vi.fn();
    bus.subscribe("record.updated", fn);
    bus.emit("record.updated", { model: "crm.lead", id: 1 });
    expect(fn).toHaveBeenCalledWith({ model: "crm.lead", id: 1 });
  });
});
