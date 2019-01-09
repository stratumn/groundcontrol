import Commit from './Commit';
import { request } from 'graphql-request';
import { Type } from 'class-transformer';
import Github from './github';
import gql from './gql';
import {
  ProjectCommitsQuery,
  ProjectCommitsQuery_repository_ref_target_Commit
} from './__generated__/ProjectCommitsQuery';

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
  name: string;
  repo: string;
  branch: string;

  id(): string {
    return this.repo
  }

  @Type(() => Commit)
  async commits({first, last}: {first?: number, last?: number}): Promise<Commit[]> {
    const parts = this.repo.split('/');

    const data = await Github.request<ProjectCommitsQuery>(commitsQuery, {
      owner: parts[1],
      repo: parts[2],
      branch: this.branch,
      first: first,
      last: last
    });

    const target = data.repository.ref.target as ProjectCommitsQuery_repository_ref_target_Commit

    return target.history.edges.map(edge => new Commit(
      edge.node.oid,
      edge.node.messageHeadline,
      edge.node.message,
      edge.node.author.name,
      new Date(edge.node.pushedDate)
    ));
  }
}

export default Project;
