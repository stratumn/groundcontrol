import PQueue from "p-queue";

import { Job, JobStatus, JobsUserArgs, Project } from "../__generated__/groundcontrol";

import pubsub, { JOB_ADDED, JOB_UPDATED } from "../pubsub";
import { connectionFromArray } from "../util/connection";
import { toGlobalId } from "./globalid";
import node from "./node";
import Type from "./type";

const allJobs: Job[] = [];
const queue = new PQueue({
  autoStart: true,
  concurrency: 2,
});

export type IFindOpts = JobsUserArgs;

export function add(name: string, project: Project, worker: () => Promise<any>): Job {
  const id = toGlobalId(Type.JOB, allJobs.length.toString(10));
  const date = new Date();
  const job: Job = {
    createdAt: date,
    id,
    name,
    project,
    status: JobStatus.Queued,
    updatedAt: date,
  };

  allJobs.unshift(job);
  node.set(id, job);

  pubsub.publish(JOB_ADDED, { jobAdded: job });

  queue.add(() => {
    job.updatedAt = new Date();
    job.status = JobStatus.Running;
    pubsub.publish(JOB_UPDATED, { jobUpdated: job });
    return worker();
  }).then(() => {
    job.updatedAt = new Date();
    job.status = JobStatus.Done;
    pubsub.publish(JOB_UPDATED, { jobUpdated: job });
  }).catch(() => {
    job.updatedAt = new Date();
    job.status = JobStatus.Failed;
    pubsub.publish(JOB_UPDATED, { jobUpdated: job });
  });

  return job;
}

export function find(opts: IFindOpts) {
  let jobs = allJobs;

  if (opts.status) {
    jobs = jobs.filter((job) => opts.status!.indexOf(job.status) >= 0);
  }

  return connectionFromArray(jobs, opts, (job) => job.id);
}

export default {
  add,
  find,
};
