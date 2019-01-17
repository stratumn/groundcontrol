import pubsub, { JOB_ADDED, JOB_UPDATED } from "../pubsub";

import { SubscriptionResolvers } from "../__generated__/groundcontrol";

const resolvers: SubscriptionResolvers.Resolvers = {
  jobAdded: {
    subscribe: () => pubsub.asyncIterator(JOB_ADDED),
  },

  jobUpdated: {
    subscribe: () => pubsub.asyncIterator(JOB_UPDATED),
  },
};

export default resolvers;
