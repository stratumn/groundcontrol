// Needs to be imported once somewhere.
import "reflect-metadata";

import cors from "cors";
import express from "express";
import graphqlHTTP from "express-graphql";
import Root from "./Root";
import schema from "./schema";
import Workspace from "./Workspace";

const app = express();

app.use(cors());
app.use("/graphql", graphqlHTTP({
  graphiql: true,
  rootValue: new Root(),
  schema,
}));

app.listen(4000);
