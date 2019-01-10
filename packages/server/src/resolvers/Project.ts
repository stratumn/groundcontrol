import gql from "graphql-tag";

import log from "../log";

import { ProjectCommitsQuery } from "../__generated__/github";

import {
  CommitConnection,
  ProjectResolvers,
} from "../__generated__/groundcontrol";

import github from "../clients/github";

// Queries the commits of a repo.
const commitsQuery = gql`
  query ProjectCommitsQuery(
    $owner: String!,
    $repo: String!,
    $branch: String!,
    $first: Int,
    $last: Int
  ) {
    repository(owner: $owner, name: $repo) {
      ref(qualifiedName: $branch) {
        target {
          ... on Commit {
            history(first: $first, last: $last) {
              pageInfo {
                hasNextPage
                hasPreviousPage
                startCursor
                endCursor
              }
              edges {
                cursor
                node {
                  oid
                  messageHeadline
                  message
                  author {
                    name
                  }
                  pushedDate
                }
              }
            }
          }
        }
      }
    }
  }
`;

const resolvers: ProjectResolvers.Resolvers = {
  id: (obj) =>
    // TODO: this isn't globally unique.
    new Buffer(`project:${obj.name}`).toString("base64"),

  commits: async (obj, { first, last }) => {
    const segments = obj.repo.split("/");

    const res = await github.query<ProjectCommitsQuery.Query, ProjectCommitsQuery.Variables>({
      query: commitsQuery,
      variables: {
        branch: obj.branch,
        first,
        last,
        owner: segments[1],
        repo: segments[2],
      },
    });

    const pageInfo = res.data.repository!.ref!.target.history.pageInfo;

    const edges = res.data.repository!.ref!.target.history.edges!.map((edge) => {
      const node = edge!.node!;

      return {
        cursor: edge!.cursor,
        node: {
          author: node.author!.name || "Unknown",
          date: node.pushedDate,
          headline: node.messageHeadline,
          id: node.oid,
          message: node.message,
        },
      };
    });

    const conn: CommitConnection = { edges, pageInfo };

    return conn;
  },
};

export default resolvers;
