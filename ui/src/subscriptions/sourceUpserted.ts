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
import { requestSubscription } from "react-relay";
import { ConnectionHandler, Environment } from "relay-runtime";

const subscription = graphql`
  subscription sourceUpsertedSubscription {
    sourceUpserted {
      ... on DirectorySource {
        ...DirectorySourceListItem_item
      }
      ... on GitSource {
        ...GitSourceListItem_item
      }
    }
  }
`;

export function subscribe(environment: Environment) {
  return requestSubscription(
    environment,
    {
      onError: (error) => console.error(error),
      subscription,
      updater: (store) => {
        const record = store.getRootField("sourceUpserted")!;
        const recordId = record.getValue("id");
        const viewer = store.getRoot().getLinkedRecord("viewer");

        const connection = ConnectionHandler.getConnection(
          viewer,
          "SourceListPage_sources",
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
            return;
          }

          const edge = ConnectionHandler.createEdge(
            store,
            connection,
            record,
            "SourcesConnection",
          );
          ConnectionHandler.insertEdgeBefore(connection, edge);
        }
    },
      variables: {},
    },
  );
}
