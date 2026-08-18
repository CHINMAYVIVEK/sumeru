import { SwcComponent } from "../../runtime/component.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { renderStubView } from "./stub-view.js";

interface StubViewProps {
  payload: SwcWorkspacePayload;
}

export class CohortView extends SwcComponent<StubViewProps> {
  template() {
    return renderStubView(this.props.payload.arch.title ?? "Cohort", this.props.payload);
  }
}
