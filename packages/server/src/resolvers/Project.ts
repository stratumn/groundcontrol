import gql from "graphql-tag";

import {
  ProjectCommitsQuery,
  ProjectDescriptionQuery,
} from "../__generated__/github";

import {
  CommitConnection,
  ProjectResolvers,
} from "../__generated__/groundcontrol";

import github from "../clients/github";

// Queries the description of a repo.
const descriptionQuery = gql`
  query ProjectDescriptionQuery(
    $owner: String!,
    $repo: String!,
  ) {
    repository(owner: $owner, name: $repo) {
      description
    }
  }
`;

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
  id: (obj, args, context, info) =>
    // TODO: this isn't globally unique.
    Buffer.from(`project:${obj.repo}`).toString("base64"),

  description: async (obj) => {
    const segments = obj.repo.split("/");

    const res = await github.query<ProjectDescriptionQuery.Query, ProjectDescriptionQuery.Variables>({
      query: descriptionQuery,
      variables: {
        owner: segments[1],
        repo: segments[2],
      },
    });

    return res.data.repository!.description;
  },

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
