import { Type } from "class-transformer";
import {
  ProjectCommitsQuery,
  ProjectCommitsQuery_repository_ref_target_Commit,
} from "./__generated__/ProjectCommitsQuery";
import Commit from "./Commit";
import Github from "./github";
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

class Project {
  constructor(
    public name: string,
    public repo: string,
    public branch: string,
  ) {
  }

  public id(): string {
    return this.repo;
  }

  @Type(() => Commit)
  public async commits({first, last}: {first?: number, last?: number}): Promise<Commit[]> {
    const parts = this.repo.split("/");

    const data = await Github.request<ProjectCommitsQuery>(commitsQuery, {
      branch: this.branch,
      first,
      last,
      owner: parts[1],
      repo: parts[2],
    });

    if (!data || !data.repository || !data.repository.ref) {
      throw new Error(`repo ${this.repo}@${this.branch} is not a valid repo`);
    }

    const target = data.repository.ref.target as ProjectCommitsQuery_repository_ref_target_Commit;

    if (!target.history.edges) {
      return [];
    }

    const commits: Commit[] = [];

    target.history.edges.forEach((edge) => {
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
  }
}

export default Project;
