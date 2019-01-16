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
  return fetch("http://localhost:4000/graphql", {
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
  const client = new SubscriptionClient("ws://localhost:4000/graphql", { reconnect: true });

  const { unsubscribe } = client
    .request({ query, variables })
    .subscribe({
      complete: onCompleted!.bind(observer),
      error: onError!.bind(observer),
      next: onNext!.bind(observer),
    });

  return { dispose: unsubscribe };
};

export default new Environment({
  network: Network.create(fetchQuery, setupSubscription),
  store: new Store(new RecordSource()),
});
