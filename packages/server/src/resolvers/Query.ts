import { QueryResolvers } from "../__generated__/groundcontrol";

import { toGlobalId } from "../models/globalid";
import jobs from "../models/job";
import node from "../models/node";
import Type from "../models/type";
import workspaces from "../models/workspace";

const resolvers: QueryResolvers.Resolvers = {
  node: (obj, { id }) => node.get(id),

  workspaces: workspaces.all,

  workspace: (obj, { slug }) => workspaces.get(toGlobalId(Type.WORKSPACE, slug)),

  jobs: (obj, args) => {
    return jobs.find(args);
  },
};

export default resolvers;
