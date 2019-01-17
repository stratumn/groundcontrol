import pubsub, { JOB_UPSERTED } from "../pubsub";

import { SubscriptionResolvers } from "../__generated__/groundcontrol";

const resolvers: SubscriptionResolvers.Resolvers = {
  jobUpserted: {
    subscribe: () => pubsub.asyncIterator(JOB_UPSERTED),
  },
};

export default resolvers;
