import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import { Card } from "semantic-ui-react";

import { WorkspaceCardGroup_items } from "./__generated__/WorkspaceCardGroup_items.graphql";

import WorkspaceCard from "./WorkspaceCard";

interface IProps {
  items: WorkspaceCardGroup_items;
  onClone: (id: string) => any;
}

export class WorkspaceCardGroup extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const cards = items.map((item) => (
      <WorkspaceCard
        key={item.id}
        item={item}
        onClone={this.handleClone.bind(this, item.id)}
      />
     ));

    return <Card.Group itemsPerRow={3}>{cards}</Card.Group>;
  }

  private handleClone(id: string) {
    this.props.onClone(id);
  }
}

export default createFragmentContainer(WorkspaceCardGroup, graphql`
  fragment WorkspaceCardGroup_items on Workspace
    @relay(plural: true) {
    ...WorkspaceCard_item
    id
  }`,
);
