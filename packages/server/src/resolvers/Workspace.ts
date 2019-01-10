import { WorkspaceResolvers } from "../__generated__/groundcontrol";

const resolvers: WorkspaceResolvers.Resolvers = {
  id: (obj) =>
    new Buffer(`workspace:${obj.slug}`).toString("base64"),
};

export default resolvers;
