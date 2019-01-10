import fs from "fs";
import { IResolvers, mergeSchemas} from "graphql-tools";
import { join } from "path";
import { promisify } from "util";

import github from "./github";
import groundcontrol from "./groundcontrol";

import githubRepo from "../resolvers/link/githubRepo";

export default async () => {

  const githubSchema = await github();
  const groundcontrolSchema = await groundcontrol();

  const filename = join(__dirname, "../../link.graphql");
  const linkSchema = await promisify(fs.readFile)(filename, { encoding: "utf8" });

  return mergeSchemas({
    resolvers: {
      Project: {
          githubRepo,
      },
    } as IResolvers,
    schemas: [
      githubSchema,
      groundcontrolSchema,
      linkSchema,
    ],
  });
};
