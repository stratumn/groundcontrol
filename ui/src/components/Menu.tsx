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
import { Link } from "found";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import { Container, Label, Menu as SemanticMenu } from "semantic-ui-react";

import { Menu_jobMetrics } from "./__generated__/Menu_jobMetrics.graphql";

interface IProps {
  jobMetrics: Menu_jobMetrics;
}

export class Menu extends Component<IProps> {
  public render() {
    const { queued, running } = this.props.jobMetrics;
    const active = queued + running;

    return (
      <SemanticMenu fixed="top" color="teal" inverted={true}>
        <Container>
          <Link
            to="/workspaces"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Workspaces
          </Link>
          <Link
            to="/jobs"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Jobs
            <Label color="blue">
              {active}
            </Label>
          </Link>
          <Link
            to="/processes"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Processes
          </Link>
          <Link
            to="/errors"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Logs
          </Link>
        </Container>
      </SemanticMenu>
    );
  }
}

export default createFragmentContainer(Menu, graphql`
  fragment Menu_jobMetrics on JobMetrics {
    queued
    running
  }`,
);
