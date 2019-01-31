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
import { createFragmentContainer } from "react-relay";
import { Card, Segment } from "semantic-ui-react";

import { ProcessGroupCardGroup_items } from "./__generated__/ProcessGroupCardGroup_items.graphql";

import ProcessGroupCard from "./ProcessGroupCard";

interface IProps {
  items: ProcessGroupCardGroup_items;
  onStartGroup: (id: string) => any;
  onStopGroup: (id: string) => any;
  onStartProcess: (id: string) => any;
  onStopProcess: (id: string) => any;
}

export class ProcessGroupCardGroup extends Component<IProps> {

  public render() {
    const items = this.props.items;

    if (items.length < 1) {
      return <Segment>There are no processes at this time.</Segment>;
    }

    const { onStartProcess, onStopProcess } = this.props;
    const cards = items.map((item) => (
      <ProcessGroupCard
        key={item.id}
        item={item}
        onStartGroup={this.handleStartGroup.bind(this, item.id)}
        onStopGroup={this.handleStopGroup.bind(this, item.id)}
        onStartProcess={onStartProcess}
        onStopProcess={onStopProcess}
      />
     ));

    return <Card.Group>{cards}</Card.Group>;
  }

  private handleStartGroup(id: string) {
    this.props.onStartGroup(id);
  }

  private handleStopGroup(id: string) {
    this.props.onStopGroup(id);
  }

}

export default createFragmentContainer(ProcessGroupCardGroup, graphql`
  fragment ProcessGroupCardGroup_items on ProcessGroup
    @relay(plural: true) {
    ...ProcessGroupCard_item
    id
  }`,
);
