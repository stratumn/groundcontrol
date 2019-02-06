// Copyright 2019 Stratumn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import graphql from "babel-plugin-relay/macro";
import { permutationCombination } from "js-combinatorics";
import { requestSubscription } from "react-relay";
import { ConnectionHandler, Environment } from "relay-runtime";

const subscription = graphql`
  subscription jobUpsertedSubscription($lastMessageId: ID) {
    jobUpserted(lastMessageId: $lastMessageId) {
      ...JobTable_items
    }
  }
`;

// Compute all possible combinations of status in order to update filtered connection.
const allStatusCombinations = permutationCombination(
  ["QUEUED", "RUNNING", "STOPPING", "DONE", "FAILED"],
).toArray() as string[][];

// Since there are many combinations we keep a map of combinations that contains a status.
//
// Note: relay has a ConnectionHandler.getConnections() method on its todo list that would be useful to avoid all of
// this.
const statusCombinations: { [s: string]: string[][] } = {
  DONE: allStatusCombinations.filter((perm: string[]) => perm.indexOf("DONE") >= 0),
  FAILED: allStatusCombinations.filter((perm: string[]) => perm.indexOf("FAILED") >= 0),
  QUEUED: allStatusCombinations.filter((perm: string[]) => perm.indexOf("QUEUED") >= 0),
  RUNNING: allStatusCombinations.filter((perm: string[]) => perm.indexOf("RUNNING") >= 0),
  STOPPING: allStatusCombinations.filter((perm: string[]) => perm.indexOf("STOPPING") >= 0),
};

// Used to remove updated jobs from connections.
const notStatusCombinations: { [s: string]: string[][] } = {
  DONE: allStatusCombinations.filter((perm: string[]) => perm.indexOf("DONE") < 0),
  FAILED: allStatusCombinations.filter((perm: string[]) => perm.indexOf("FAILED") < 0),
  QUEUED: allStatusCombinations.filter((perm: string[]) => perm.indexOf("QUEUED") < 0),
  RUNNING: allStatusCombinations.filter((perm: string[]) => perm.indexOf("RUNNING") < 0),
  STOPPING: allStatusCombinations.filter((perm: string[]) => perm.indexOf("STOPPING") < 0),
};

export function subscribe(environment: Environment, lastMessageId?: string) {
  return requestSubscription(
    environment,
    {
      onError: (error) => console.error(error),
      subscription,
      updater: (store) => {
        const record = store.getRootField("jobUpserted")!;
        const recordId = record.getValue("id");
        const status = record!.getValue("status");
        const system = store.getRoot().getLinkedRecord("system");

        // Remove job from connections that don't have the new status.
        for (const combination of notStatusCombinations[status]) {
          const connection = ConnectionHandler.getConnection(
            system,
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
            system,
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
      variables: {
        lastMessageId,
      },
    },
  );
}
