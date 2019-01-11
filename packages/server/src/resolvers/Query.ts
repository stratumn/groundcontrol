import { QueryResolvers } from "../__generated__/groundcontrol";

import { all, get } from "../data/workspaces";

const resolvers: QueryResolvers.Resolvers = {
  workspaces: all,

  workspace: async (obj, { slug }) => await get(slug),
};

export default resolvers;
