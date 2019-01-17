import graphql from "babel-plugin-relay/macro";
import { requestSubscription } from "react-relay";
import { ConnectionHandler } from "relay-runtime";

import groundcontrol from "../groundcontrol.env.relay";

const subscription = graphql`
  subscription jobAddedSubscription {
    jobAdded {
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
      updater: (store) => {
        const record = store.getRootField("jobAdded");
        const status = record!.getValue("status");
        const viewer = store.getRoot().getLinkedRecord("viewer");

        const connection = ConnectionHandler.getConnection(
          viewer,
          "JobsPage_jobs",
        );

        if (connection) {
          const edge = ConnectionHandler.createEdge(
            store,
            connection,
            record,
            "JobsConnection",
          );
          ConnectionHandler.insertEdgeBefore(connection, edge);
        }
    },
      variables: {},
    },
  );
}
