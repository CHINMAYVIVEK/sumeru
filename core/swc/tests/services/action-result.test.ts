import { describe, it, expect } from "vitest";
import { ActionService } from "../../src/services/action.js";

describe("action call results", () => {
  it("parses wizard redirects into dialog opens", async () => {
    const action = new ActionService();
    const parsed = (
      action as unknown as {
        parseRedirectOpen: (url: string) => {
          model: string;
          recordId: number;
          target: string;
        } | null;
      }
    ).parseRedirectOpen("/web?model=crm.lead.lost&id=4&view_type=form");
    expect(parsed?.model).toBe("crm.lead.lost");
    expect(parsed?.recordId).toBe(4);
    expect(parsed?.target).toBe("dialog");
  });
});
