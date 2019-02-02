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

import {
  Environment,
  Network,
  RecordSource,
  RequestNode,
  Store,
  SubscribeFunction,
  Variables,
} from "relay-runtime";
import { SubscriptionClient } from "subscriptions-transport-ws";

async function fetchQuery(operation: RequestNode, variables: Variables) {
  return fetch("http://localhost:3333/query", {
    body: JSON.stringify({
      query: operation.text,
      variables,
    }),
    headers: {
      "Content-Type": "application/json",
    },
    method: "POST",
  }).then((response) => {
    return response.json();
  });
}

const setupSubscription: SubscribeFunction = (config, variables, _, observer) => {
  const query = config.text;
  const { onNext, onError, onCompleted } = observer;
  const client = new SubscriptionClient("ws://localhost:3333/query", { reconnect: true });

  const { unsubscribe } = client
    .request({ query, variables })
    .subscribe({
      complete: onCompleted,
      error: onError,
      next: onNext,
    });

  return { dispose: unsubscribe };
};

export default new Environment({
  network: Network.create(fetchQuery, setupSubscription),
  store: new Store(new RecordSource()),
});
