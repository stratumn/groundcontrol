import { WorkspaceResolvers } from "../__generated__/groundcontrol";

import { isCloned } from "../models/project";

const resolvers: WorkspaceResolvers.Resolvers = {
  isCloning: async ({ projects }) => {
    for (const project of projects) {
      if (project.isCloning) {
        return true;
      }
    }

    return false;
  },

  isCloned: async ({ projects }) => {
    for (const project of projects) {
      if (!(await isCloned(project))) {
        return false;
      }

    }

    return true;
  },
};

export default resolvers;
