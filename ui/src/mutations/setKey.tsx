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
import { commitMutation } from "react-relay";
import { Environment } from "relay-runtime";

import { KeyInput } from "../mutations/__generated__/setKeyMutation.graphql";

const mutation = graphql`
  mutation setKeyMutation($input: KeyInput!) {
    setKey(input: $input) {
      id
    }
  }
`;

export function commit(environment: Environment, input: KeyInput) {
  commitMutation(environment, {
    mutation,
    variables: { input },
  });
}
