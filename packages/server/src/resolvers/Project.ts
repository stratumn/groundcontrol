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
import { ownerAndName } from "../util/repository";

// Queries the description of a repo.
const descriptionQuery = gql`
  query ProjectDescriptionQuery(
    $owner: String!
    $name: String!
  ) {
    repository(owner: $owner, name: $name) {
      description
    }
  }
`;

// Queries the commits of a repo.
const commitsQuery = gql`
  query ProjectCommitsQuery(
    $owner: String!
    $name: String!
    $branch: String!
    $before: String
    $after: String
    $first: Int
    $last: Int
  ) {
    repository(owner: $owner, name: $name) {
      ref(qualifiedName: $branch) {
        target {
          ... on Commit {
            history(before: $before, after: $after, first: $first, last: $last) {
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
  description: async ({ repository }) => {
    const [owner, name] = ownerAndName(repository);

    const res = await github.query<ProjectDescriptionQuery.Query, ProjectDescriptionQuery.Variables>({
      query: descriptionQuery,
      variables: {
        name,
        owner,
      },
    });

    return res.data.repository!.description;
  },

  commits: async ({ branch, repository }, { before, after, first, last }) => {
    const [owner, name] = ownerAndName(repository);

    const res = await github.query<ProjectCommitsQuery.Query, ProjectCommitsQuery.Variables>({
      query: commitsQuery,
      variables: {
        after,
        before,
        branch,
        first,
        last,
        name,
        owner,
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
