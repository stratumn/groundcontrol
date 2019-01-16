import { ApolloServer } from "apollo-server-express";
import cors from "cors";
import express from "express";

import log from "./log";

import workspaces from "./models/workspace";
import schema from "./schema";

(async () => {
  // Force an initial load.
  await workspaces.all();

  const server = new ApolloServer({
    schema: await schema(),
    tracing: process.env.APOLLO_TRACING === "1",
  });

  const app = express();
  app.use(cors());
  server.applyMiddleware({ app });

  const port = 4000;

  app.listen({ port }, () =>
    log.info(`🚀 Server ready at http://localhost:${port}${server.graphqlPath}`),
  );
})();
