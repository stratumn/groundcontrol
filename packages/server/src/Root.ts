import Workspace from "./Workspace";
import { getWorkspaces } from "./workspaces";

class Root {
  public workspaces(): Workspace[] {
    return getWorkspaces();
  }
}

export default Root;
