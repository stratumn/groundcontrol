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
import React, { Component } from "react";
import { createFragmentContainer, RelayProp } from "react-relay";
import { Disposable } from "relay-runtime";
import { Container } from "semantic-ui-react";

import { App_system } from "./__generated__/App_system.graphql";

import Menu from "../components/Menu";
import { subscribe as subscribeJobMetrics } from "../subscriptions/jobMetricsUpdated";
import { subscribe as subscribeLogMetrics } from "../subscriptions/logMetricsUpdated";
import { subscribe as subscribeProcessMetrics } from "../subscriptions/processMetricsUpdated";

import "./App.css";

interface IProps {
  relay: RelayProp;
  system: App_system;
}

export class App extends Component<IProps> {

  private disposables: Disposable[] = [];

  public render() {
    const system = this.props.system;

    return (
      <div className="App">
        <Menu system={system} />
        <Container fluid={true}>
          {this.props.children}
        </Container>
      </div>
    );
  }

  public componentDidMount() {
    const environment = this.props.relay.environment;
    const lastMessageId = this.props.system.lastMessageId;
    this.disposables.push(subscribeJobMetrics(environment, lastMessageId));
    this.disposables.push(subscribeProcessMetrics(environment, lastMessageId));
    this.disposables.push(subscribeLogMetrics(environment, lastMessageId));
  }

  public componentWillUnmount() {
    for (const disposable of this.disposables) {
      disposable.dispose();
    }

    this.disposables = [];
  }
}

export default createFragmentContainer(App, graphql`
  fragment App_system on System {
    lastMessageId
    ...Menu_system
  }`,
);
