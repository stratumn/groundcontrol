import graphql from "babel-plugin-relay/macro";
import { requestSubscription } from "react-relay";
import { ConnectionHandler } from "relay-runtime";

import groundcontrol from "../groundcontrol.env.relay";

const subscription = graphql`
  subscription jobUpdatedSubscription {
    jobUpdated {
      ...JobsList_items
    }
  }
`;

export default function() {
  return requestSubscription(
    groundcontrol,
    {
      onError: (error) => console.error(error),
      subscription,
      variables: {},
    },
  );
}
