import fs from "fs";
import { IResolvers, makeExecutableSchema } from "graphql-tools";
import { join } from "path";
import { promisify } from "util";

import Project from "./resolvers/Project";
import Query from "./resolvers/Query";
import Workspace from "./resolvers/Workspace";

export default async () => {
  const filename = join(__dirname, "../schema.graphql");
  const typeDefs = await promisify(fs.readFile)(filename, { encoding: "utf8" });

  const resolvers: IResolvers = {
    Project: Project as IResolvers,
    Query: Query as IResolvers,
    Workspace: Workspace as IResolvers,
  };

  const schema = makeExecutableSchema({
    resolvers,
    typeDefs,
  });

  return schema;
};
