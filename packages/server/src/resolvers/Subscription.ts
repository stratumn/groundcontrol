import { withFilter } from "apollo-server";

import pubsub, { JOB_UPSERTED, PROJECT_UPDATED, WORKSPACE_UPDATED } from "../pubsub";

import { SubscriptionResolvers } from "../__generated__/groundcontrol";

const resolvers: SubscriptionResolvers.Resolvers = {
  jobUpserted: {
    subscribe: () => pubsub.asyncIterator(JOB_UPSERTED),
  },

  projectUpdated: {
    subscribe: withFilter(
      () => pubsub.asyncIterator(PROJECT_UPDATED),
      (payload, variables) => !variables.id || variables.id === payload.projectUpdated.id,
    ),
  },

  workspaceUpdated: {
    subscribe: withFilter(
      () => pubsub.asyncIterator(WORKSPACE_UPDATED),
      (payload, variables) => !variables.id || variables.id === payload.workspaceUpdated.id,
    ),
  },
};

export default resolvers;
