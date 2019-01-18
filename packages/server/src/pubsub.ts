import { PubSub } from "apollo-server";

export const JOB_UPSERTED = "JOB_UPSERTED";
export const PROJECT_UPDATED = "PROJECT_UPDATED";
export const WORKSPACE_UPDATED = "WORKSPACE_UPDATED";

export default new PubSub();
