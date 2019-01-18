import graphql from "babel-plugin-relay/macro";
import { permutationCombination } from "js-combinatorics";
import { requestSubscription } from "react-relay";
import { ConnectionHandler, Environment } from "relay-runtime";

const subscription = graphql`
  subscription jobUpsertedSubscription {
    jobUpserted {
      ...JobList_items
    }
  }
`;

// Compute all possible combinations of status in order to update filtered connection.
const allStatusCombinations = permutationCombination(["QUEUED", "RUNNING", "DONE", "FAILED"]).toArray() as string[][];

// Since there are many combinations we keep a map of combinations that contains a status.
//
// Note: relay has a ConnectionHandler.getConnections() method on its todo list that would be useful to avoid all of
// this.
const statusCombinations: { [s: string]: string[][] } = {
  DONE: allStatusCombinations.filter((perm: string[]) => perm.indexOf("DONE") >= 0),
  FAILED: allStatusCombinations.filter((perm: string[]) => perm.indexOf("FAILED") >= 0),
  QUEUED: allStatusCombinations.filter((perm: string[]) => perm.indexOf("QUEUED") >= 0),
  RUNNING: allStatusCombinations.filter((perm: string[]) => perm.indexOf("RUNNING") >= 0),
};

// Used to remove updated jobs from connections.
const notStatusCombinations: { [s: string]: string[][] } = {
  DONE: allStatusCombinations.filter((perm: string[]) => perm.indexOf("DONE") < 0),
  FAILED: allStatusCombinations.filter((perm: string[]) => perm.indexOf("FAILED") < 0),
  QUEUED: allStatusCombinations.filter((perm: string[]) => perm.indexOf("QUEUED") < 0),
  RUNNING: allStatusCombinations.filter((perm: string[]) => perm.indexOf("RUNNING") < 0),
};

export function subscribe(environment: Environment) {
  return requestSubscription(
    environment,
    {
      onError: (error) => console.error(error),
      subscription,
      updater: (store) => {
        const record = store.getRootField("jobUpserted")!;
        const recordId = record.getValue("id");
        const status = record!.getValue("status");
        const viewer = store.getRoot().getLinkedRecord("viewer");

        // Remove job from connections that don't have the new status.
        for (const combination of notStatusCombinations[status]) {
          const connection = ConnectionHandler.getConnection(
            viewer,
            "JobListPage_jobs",
            { status: combination },
          );

          if (connection) {
            ConnectionHandler.deleteNode(connection, recordId);
          }
        }

        // Add job to connections that have the new status (if it doesn't already exist).
        for (const combination of [undefined, ...statusCombinations[status]]) {
          const connection = ConnectionHandler.getConnection(
            viewer,
            "JobListPage_jobs",
            { status: combination },
          );

          if (connection) {
            const edges = connection.getLinkedRecords("edges");
            let exists = false;

            for (const e of edges) {
              const id = e.getLinkedRecord("node")!.getValue("id");

              if (recordId === id) {
                exists = true;
                break;
              }
            }

            if (exists) {
              continue;
            }

            const edge = ConnectionHandler.createEdge(
              store,
              connection,
              record,
              "JobsConnection",
            );
            ConnectionHandler.insertEdgeBefore(connection, edge);
          }
        }
    },
      variables: {},
    },
  );
}
