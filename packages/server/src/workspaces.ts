import { join } from 'path';
import Workspace from './Workspace';

let workspaces: Workspace[];

function reload(): Workspace[] {
  workspaces = Workspace.load(join(__dirname, '../../../groundcontrol.yml'));

  return workspaces;
}

export function getWorkspaces(): Workspace[] {
  if (workspaces == null) {
    return reload();
  }

  return workspaces;
}
