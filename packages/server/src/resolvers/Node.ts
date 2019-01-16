import { NodeResolvers } from "../__generated__/groundcontrol";

import { fromGlobalId } from "../models/globalid";

const resolvers: NodeResolvers.Resolvers = {
  __resolveType(obj) {
    return fromGlobalId(obj.id)[0];
  },
};

export default resolvers;
