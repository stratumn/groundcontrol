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
import { Environment } from "relay-runtime";

const subscription = graphql`
  subscription jobMetricsUpdatedSubscription {
    jobMetricsUpdated {
      ...Menu_jobMetrics
    }
  }
`;

export function subscribe(environment: Environment) {
  return requestSubscription(
    environment,
    {
      onError: (error) => console.error(error),
      subscription,
      variables: {},
    },
  );
}
