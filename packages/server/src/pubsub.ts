import { PubSub } from "apollo-server";

export const JOB_ADDED = "JOB_ADDED";
export const JOB_UPDATED = "JOB_UPDATED";

export default new PubSub();
