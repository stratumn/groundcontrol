import Workspace from './Workspace';
import { getWorkspaces } from './workspaces';

class Root {
  workspaces():Workspace[] {
    return getWorkspaces();
  }
};

export default Root;
