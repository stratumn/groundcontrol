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
import { Card, SemanticWIDTHS } from "semantic-ui-react";

import { WorkspaceCardGroup_items } from "./__generated__/WorkspaceCardGroup_items.graphql";

import WorkspaceCard from "./WorkspaceCard";

interface IProps {
  items: WorkspaceCardGroup_items;
  itemsPerRow: SemanticWIDTHS;
  onClone: (id: string) => any;
  onPull: (id: string) => any;
}

export class WorkspaceCardGroup extends Component<IProps> {

  public render() {
    const { items, itemsPerRow } = this.props;
    const cards = items.map((item) => (
      <WorkspaceCard
        key={item.id}
        item={item}
        onClone={this.handleClone.bind(this, item.id)}
        onPull={this.handlePull.bind(this, item.id)}
      />
     ));

    return (
      <Card.Group itemsPerRow={itemsPerRow}>
        {cards}
      </Card.Group>
    );
  }

  private handleClone(id: string) {
    this.props.onClone(id);
  }

  private handlePull(id: string) {
    this.props.onPull(id);
  }
}

export default createFragmentContainer(WorkspaceCardGroup, graphql`
  fragment WorkspaceCardGroup_items on Workspace
    @relay(plural: true) {
    ...WorkspaceCard_item
    id
  }`,
);
