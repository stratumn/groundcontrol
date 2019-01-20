import graphql from "babel-plugin-relay/macro";
import { commitMutation } from "react-relay";
import { Environment } from "relay-runtime";

const mutation = graphql`
  mutation cloneProjectMutation($id: String!) {
    cloneProject(id: $id) {
      id
    }
  }
`;

export function commit(environment: Environment, id: string) {
  commitMutation(environment, {
    mutation,
    variables: { id },
  });
}
