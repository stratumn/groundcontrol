/*import { Type } from "class-transformer";
import { ProjectCommitsQuery } from "./__generated__/github";
import {
  CommitConnection,
  ProjectResolvers,
 } from "./__generated__/groundcontrol";
import Commit from "./Commit";
import Github from "./schemas/github";
import gql from "./gql";

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
              edges {
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
    new Buffer(`project:${obj.repo}`).toString("base64"),

  name: (obj) => obj.name,

  commits: (parent, { first, last }) => {
    return null;
    const parts = this.repo.split("/");

    const data = await Github.request<ProjectCommitsQuery.Query>(commitsQuery, {
      branch: this.branch,
      first,
      last,
      owner: parts[1],
      repo: parts[2],
    });

    if (!data || !data.repository || !data.repository.ref) {
      throw new Error(`repo ${this.repo}@${this.branch} is not a valid repo`);
    }

    if (!data.repository.ref.target.history.edges) {
      return [];
    }

    const commits: Commit[] = [];

    data.repository.ref.target.history.edges.forEach((edge) => {
      if (!edge || !edge.node) {
        return;
      }

      commits.push(new Commit(
        edge.node.oid,
        edge.node.messageHeadline,
        edge.node.message,
        edge.node.author && edge.node.author.name ? edge.node.author.name : "Unknown",
        new Date(edge.node.pushedDate),
      ));
    });

    return commits;
  },
};

export default resolvers;/*
