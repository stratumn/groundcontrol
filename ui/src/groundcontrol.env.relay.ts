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
  return fetch("http://localhost:8080/query", {
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
  const client = new SubscriptionClient("ws://localhost:8080/query", { reconnect: true });

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
