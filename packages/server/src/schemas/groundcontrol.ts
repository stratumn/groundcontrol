import fs from "fs";
import { IResolvers, makeExecutableSchema } from "graphql-tools";
import { join } from "path";
import { promisify } from "util";

import Query from "../resolvers/groundcontrol/Query";
import Workspace from "../resolvers/groundcontrol/Workspace";

export default async () => {
  const filename = join(__dirname, "../../groundcontrol.graphql");
  const typeDefs = await promisify(fs.readFile)(filename, { encoding: "utf8" });

  const resolvers: IResolvers = {
    Query: Query as IResolvers,
    Workspace: Workspace as IResolvers,
  };

  const schema = makeExecutableSchema({
    resolvers,
    typeDefs,
  });

  return schema;
};
