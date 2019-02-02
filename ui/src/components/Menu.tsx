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

import { Menu_system } from "./__generated__/Menu_system.graphql";

interface IProps {
  system: Menu_system;
}

export class Menu extends Component<IProps> {
  public render() {
    const { jobMetrics, processMetrics, logMetrics } = this.props.system;

    return (
      <SemanticMenu
        fixed="top"
        size="large"
        color="teal"
        inverted={true}
      >
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
              {jobMetrics.queued + jobMetrics.running}
            </Label>
          </Link>
          <Link
            to="/processes"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Processes
            <Label color="blue">
              {processMetrics.running}
            </Label>
          </Link>
          <Link
            to="/logs"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Logs
            <Label color="pink">
              {logMetrics.error}
            </Label>
          </Link>
          <SemanticMenu.Item href="http://localhost:3333/graphql">
            GraphQL
          </SemanticMenu.Item>
        </Container>
      </SemanticMenu>
    );
  }
}

export const jobMetricsFragment = graphql`
  fragment Menu_jobMetrics on JobMetrics {
    queued
    running
  }
`;

export const processMetricsFragment = graphql`
  fragment Menu_processMetrics on ProcessMetrics {
    running
  }
`;

export const logMetricsFragment = graphql`
  fragment Menu_logMetrics on LogMetrics {
    error
  }
`;

export default createFragmentContainer(Menu, graphql`
  fragment Menu_system on System {
    jobMetrics {
      ...Menu_jobMetrics @relay(mask: false)
    }
    processMetrics {
      ...Menu_processMetrics @relay(mask: false)
    }
    logMetrics {
      ...Menu_logMetrics @relay(mask: false)
    }
  }`,
);
