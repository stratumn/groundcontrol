import { ensureDir } from "fs-extra";
import { join } from "path";
import simpleGit from "simple-git/promise";

import { Project } from "../__generated__/groundcontrol";

import { ownerAndRepo } from "../util/repo";
import jobs from "./job";
import node from "./node";
import { workspacesRoot } from "./workspace";

const git = simpleGit();

export function get(gid: string) {
  return node.get(gid) as Project;
}

export function clone(gid: string) {
  const project = get(gid);

  return jobs.add(
    `Clone "${project.repo}@${project.branch}" into workspace "${project.workspace.name}"`,
    project,
    async () => {
      const [ owner, repo ] = ownerAndRepo(project.repo);
      const localParentPath = join(workspacesRoot, project.workspace.slug, owner);
      const localPath = join(localParentPath, repo);

      await ensureDir(localParentPath);
      await git.clone(project.repo, localPath);
      await git.checkout(project.branch);
    },
  );
}

export default {
  clone,
  get,
};
