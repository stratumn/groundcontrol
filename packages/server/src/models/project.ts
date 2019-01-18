import { ensureDir } from "fs-extra";
import { join } from "path";
import simpleGit from "simple-git/promise";

import { Project } from "../__generated__/groundcontrol";

import log from "../log";

import pubsub, { PROJECT_UPDATED, WORKSPACE_UPDATED } from "../pubsub";
import { ownerAndName } from "../util/repository";
import jobs from "./job";
import node from "./node";
import { workspacesRoot } from "./workspace";

export function get(gid: string) {
  return node.get(gid) as Project;
}

export function clone(gid: string) {
  const project = get(gid);

  if (project.isCloning) {
    throw new Error(`Project "${project.repository}" is already being clone`);
  }

  project.isCloning = true;
  publishProjectUpdated(project);

  return jobs.add(
    `Clone "${project.repository}@${project.branch}" into workspace "${project.workspace.name}"`,
    project,
    async () => {
      let err: Error | undefined;

      try {
        const git = simpleGit();
        const [ owner, name ] = ownerAndName(project.repository);
        const localParentPath = join(workspacesRoot, project.workspace.slug, owner);
        const localPath = join(localParentPath, name);

        await ensureDir(localParentPath);
        await git.clone(project.repository, localPath);
        await git.checkout(project.branch);
      } catch (e) {
        err = e;
      }

      project.isCloning = false;
      publishProjectUpdated(project);

      if (err) {
        throw err;
      }
    },
  );
}

export async function isCloned(project: Project) {
  const [ owner, name ] = ownerAndName(project.repository);
  const localParentPath = join(workspacesRoot, project.workspace.slug, owner);
  const localPath = join(localParentPath, name);

  try {
    const git = simpleGit(localPath); // throws if not a dir
    const isRepository = await git.checkIsRepo();
    return isRepository;
  } catch (e) {
    log.error(e);
    return false;
  }
}

function publishProjectUpdated(project: Project) {
  pubsub.publish(PROJECT_UPDATED, { projectUpdated: project });
  pubsub.publish(WORKSPACE_UPDATED, { workspaceUpdated: project.workspace });
}

export default {
  clone,
  get,
  isCloned,
};
