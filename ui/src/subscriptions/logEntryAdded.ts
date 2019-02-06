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
  subscription logEntryAddedSubscription($lastMessageId: ID) {
    logEntryAdded(lastMessageId: $lastMessageId) {
      ...LogEntryTable_items
    }
  }
`;

// Compute all possible combinations of level in order to update filtered connection.
const allLevelCombinations = permutationCombination(["DEBUG", "INFO", "WARNING", "ERROR"]).toArray() as string[][];

// Since there are many combinations we keep a map of combinations that contains a level.
//
// Note: relay has a ConnectionHandler.getConnections() method on its todo list that would be useful to avoid all of
// this.
const levelCombinations: { [s: string]: string[][] } = {
  DEBUG: allLevelCombinations.filter((perm: string[]) => perm.indexOf("DEBUG") >= 0),
  ERROR: allLevelCombinations.filter((perm: string[]) => perm.indexOf("ERROR") >= 0),
  INFO: allLevelCombinations.filter((perm: string[]) => perm.indexOf("INFO") >= 0),
  WARNING: allLevelCombinations.filter((perm: string[]) => perm.indexOf("WARNING") >= 0),
};

export function subscribe(environment: Environment, lastMessageId?: string) {
  return requestSubscription(
    environment,
    {
      onError: (error) => console.error(error),
      subscription,
      updater: (store) => {
        const record = store.getRootField("logEntryAdded")!;
        const recordId = record.getValue("id");
        const level = record!.getValue("level");
        const system = store.getRoot().getLinkedRecord("system");

        // Add log entry to connections that have the level.
        for (const combination of [undefined, ...levelCombinations[level]]) {
          const connections = [
            ConnectionHandler.getConnection(
              system,
              "LogEntryListPage_logEntries",
              { level: combination },
            ),
          ];

          const owner = record.getLinkedRecord("owner");

          if (owner) {
            const ownerId = owner.getValue("id");
            connections.push(
              ConnectionHandler.getConnection(
                system,
                "LogEntryListPage_logEntries",
                { level: combination, ownerId },
              ),
            );
          }

          for (const connection of connections) {
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
                "LogEntrysConnection",
              );
              ConnectionHandler.insertEdgeBefore(connection, edge);
            }
          }
        }
    },
      variables: {
        lastMessageId,
      },
    },
  );
}
