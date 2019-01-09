import { GraphQLClient } from "graphql-request";

const endpoint = "https://api.github.com/graphql";

export default new GraphQLClient(endpoint, {
  headers: {
    "Authorization": `bearer ${process.env.GITHUB_TOKEN}`,
    "Content-Type": "application/json",
  },
});
