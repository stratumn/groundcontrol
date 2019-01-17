import { UserResolvers } from "../__generated__/groundcontrol";

import { toGlobalId } from "../models/globalid";
import job from "../models/job";
import type from "../models/type";
import workspace from "../models/workspace";

const resolvers: UserResolvers.Resolvers = {
  workspaces: workspace.all,

  workspace: (_, { slug }) => workspace.get(toGlobalId(type.WORKSPACE, slug)),

  jobs: (_, args) => {
    return job.find(args);
  },
};

export default resolvers;
