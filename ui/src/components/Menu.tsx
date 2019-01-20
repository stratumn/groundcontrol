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

import { Link } from "found";
import React, { Component } from "react";
import { Container, Label, Menu as SemanticMenu } from "semantic-ui-react";

export default class Menu extends Component {
  public render() {
    return (
      <SemanticMenu fixed="top" color="teal" inverted={true}>
        <Container>
          <Link
            to="/workspaces"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Workspaces
            <Label color="purple">1</Label>
          </Link>
          <Link
            to="/jobs"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Jobs
            <Label color="blue">3</Label>
          </Link>
          <Link
            to="/processes"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Processes
            <Label color="blue">1</Label>
          </Link>
          <Link
            to="/errors"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Logs
            <Label color="pink">2</Label>
          </Link>
        </Container>
      </SemanticMenu>
    );
  }
}
