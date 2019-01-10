// Needs to be imported once somewhere.
import "reflect-metadata";

import { ApolloServer } from "apollo-server-express";
import cors from "cors";
import express from "express";
import { makeExecutableSchema, mergeSchemas } from "graphql-tools";

import log from "./log";

import schema from "./schemas/link";

(async () => {
  const server = new ApolloServer({ schema: await schema() });

  const app = express();
  app.use(cors());
  server.applyMiddleware({ app });

  const port = 4000;

  app.listen({ port }, () =>
    log.info(`🚀 Server ready at http://localhost:${port}${server.graphqlPath}`),
  );
})();
