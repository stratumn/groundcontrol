import { readFile } from "fs";
import yaml from "js-yaml";
import { homedir } from "os";
import { join } from "path";
import { promisify } from "util";

import { Workspace } from "../__generated__/groundcontrol";

import { toGlobalId } from "./globalid";
import node, { set } from "./node";
import project from "./project";
import Type from "./type";

const filename = join(__dirname, "../../../../groundcontrol.yml");

export const workspacesRoot = process.env.GROUNDCONTROL_WORKSPACES_ROOT ||
  join(homedir(), "groundcontrol", "workspaces");

export async function all(): Promise<Workspace[]> {
  const data = await promisify(readFile)(filename, { encoding: "utf8" });
  const workspaces: Workspace[] = yaml.safeLoad(data).workspaces;

  for (const workspace of workspaces) {
    workspace.id = toGlobalId(Type.WORKSPACE, workspace.slug);

    node.set(workspace.id, workspace);

    for (const prj of workspace.projects!) {
      prj.id = toGlobalId(Type.PROJECT, workspace.slug, prj.repository, prj.branch);
      prj.workspace = workspace;
      prj.isCloning = false;

      node.set(prj.id, prj);
    }
  }

  return workspaces;
}

export function get(gid: string) {
  return node.get(gid) as Workspace;
}

export async function clone(gid: string) {
  const workspace = get(gid);

  return workspace!.projects!
    .filter(async (prj) => !prj.isCloning && !(await project.isCloned(prj)))
    .map(({ id }) => project.clone(id));
}

export default {
  all,
  clone,
  get,
  workspacesRoot,
};
