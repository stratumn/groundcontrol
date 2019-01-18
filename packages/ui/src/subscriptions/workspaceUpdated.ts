import graphql from "babel-plugin-relay/macro";
import { requestSubscription } from "react-relay";
import { Environment } from "relay-runtime";

import groundcontrol from "../groundcontrol.env.relay";

const subscription = graphql`
  subscription workspaceUpdatedSubscription($id: ID) {
    workspaceUpdated(id: $id) {
      ...WorkspaceCard_item
      ...WorkspaceMenu_workspace
      projects {
        ...ProjectCard_item
      }
    }
  }
`;

export function subscribe(environment: Environment, id?: string) {
  return requestSubscription(
    environment,
    {
      onError: (error) => console.error(error),
      subscription,
      variables: { id },
    },
  );
}
