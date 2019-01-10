import { InMemoryCache } from "apollo-cache-inmemory";
import { ApolloClient } from "apollo-client";
import { setContext } from "apollo-link-context";
import { HttpLink } from "apollo-link-http";
import fetch from "isomorphic-fetch";

const http = new HttpLink({ uri: "https://api.github.com/graphql", fetch });

const link = setContext((request, previousContext) => ({
  headers: {
    Authorization: `bearer ${process.env.GITHUB_TOKEN}`,
  },
})).concat(http);

const cache = new InMemoryCache();

export default new ApolloClient({ cache, link});
