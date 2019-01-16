import { ApolloServer } from "apollo-server-express";
import cors from "cors";
import express from "express";
import { createServer } from "http";
import { SubscriptionServer } from "subscriptions-transport-ws";

import log from "./log";

import workspaces from "./models/workspace";
import schema from "./schema";

const port = 4000;

(async () => {
  const skema = await schema();

  // Force an initial load.
  await workspaces.all();

  const server = new ApolloServer({
    schema: skema,
    tracing: process.env.APOLLO_TRACING === "1",
  });

  const app = express();
  app.use(cors());
  server.applyMiddleware({ app });

  const httpServer = createServer(app);
  server.installSubscriptionHandlers(httpServer);

  httpServer.listen(port, () => {
    log.info(`🚀 Server ready at http://localhost:${port}${server.graphqlPath}`);
  });
})();
