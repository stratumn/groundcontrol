import fs from "fs";
import { IResolvers, makeExecutableSchema } from "graphql-tools";
import { join } from "path";
import { promisify } from "util";

import Mutation from "./resolvers/Mutation";
import Node from "./resolvers/Node";
import Project from "./resolvers/Project";
import Query from "./resolvers/Query";
import Subscription from "./resolvers/Subscription";

export default async () => {
  const filename = join(__dirname, "../schema.graphql");
  const typeDefs = await promisify(fs.readFile)(filename, { encoding: "utf8" });

  const resolvers: IResolvers = {
    Mutation: Mutation as IResolvers,
    Node,
    Project: Project as IResolvers,
    Query: Query as IResolvers,
    Subscription: Subscription as IResolvers,
  };

  const schema = makeExecutableSchema({
    resolvers,
    typeDefs,
  });

  return schema;
};
