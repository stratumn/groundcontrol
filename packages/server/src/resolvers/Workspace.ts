import { WorkspaceResolvers } from "../__generated__/groundcontrol";

const resolvers: WorkspaceResolvers.Resolvers = {
  id: (obj) =>
    Buffer.from(`workspace:${obj.slug}`).toString("base64"),
};

export default resolvers;
