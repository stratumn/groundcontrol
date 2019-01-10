import { QueryResolvers, Workspace } from "../../__generated__/groundcontrol";

import { all } from "../../data/workspaces";

const resolvers: QueryResolvers.Resolvers = {
  workspaces: all,
};

export default resolvers;
