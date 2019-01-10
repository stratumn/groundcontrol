import { setContext } from "apollo-link-context";
import { HttpLink } from "apollo-link-http";
import { introspectSchema, makeRemoteExecutableSchema } from "graphql-tools";
import fetch from "isomorphic-fetch";

const http = new HttpLink({ uri: "https://api.github.com/graphql", fetch });

const link = setContext((request, previousContext) => ({
  headers: {
    Authorization: `bearer ${process.env.GITHUB_TOKEN}`,
  },
})).concat(http);

export default async () => {
  const schema = await introspectSchema(link);

  const executableSchema = makeRemoteExecutableSchema({
    link,
    schema,
  });

  return executableSchema;
};
