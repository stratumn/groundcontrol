import { QueryResolvers } from "../__generated__/groundcontrol";

import { toGlobalId } from "../models/globalid";
import node from "../models/node";
import type from "../models/type";

const resolvers: QueryResolvers.Resolvers = {
  node: (_, { id }) => node.get(id),

  viewer: () => ({
    id: toGlobalId(type.USER, "0"),
  }),
};

export default resolvers;
