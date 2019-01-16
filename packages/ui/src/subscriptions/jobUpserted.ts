import graphql from "babel-plugin-relay/macro";
import { requestSubscription } from "react-relay";

import groundcontrol from "../groundcontrol.env.relay";

const subscription = graphql`
  subscription jobUpsertedSubscription {
    jobUpserted {
      id
      name
      status
      createdAt
      updatedAt
      project {
        repo
        branch
        workspace {
          slug
          name
        }
      }
    }
  }
`;

export default function() {
  requestSubscription(
    groundcontrol,
    {
      onCompleted: () => {/* server closed the subscription */},
      onError: (error) => console.error(error),
      subscription,
      variables: {},
    },
  );
}
