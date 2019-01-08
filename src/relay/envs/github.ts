import {
  Environment,
  Network,
  RecordSource,
  RequestNode,
  Store,
  Variables
} from 'relay-runtime';

const fetchQuery = async (
	operation: RequestNode,
	variables: Variables,
) => {
  return fetch('https://api.github.com/graphql', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `bearer ${process.env.REACT_APP_GITHUB_TOKEN}`
    },
    body: JSON.stringify({
      query: operation.text,
      variables,
    }),
  }).then(response => {
    return response.json();
  });
}

export default new Environment({
  network: Network.create(fetchQuery),
  store: new Store(new RecordSource()),  
});
