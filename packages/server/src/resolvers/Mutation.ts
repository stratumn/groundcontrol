import { MutationResolvers } from "../__generated__/groundcontrol";

import project from "../models/project";
import workspace from "../models/workspace";

const resolvers: MutationResolvers.Resolvers = {
  cloneProject: (_, args) => project.clone(args.id),

  cloneWorkspace: (_, args) => workspace.clone(args.id),
};

export default resolvers;
