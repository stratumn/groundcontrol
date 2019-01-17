import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { commitMutation, createFragmentContainer, RelayProp } from "react-relay";
import { Card } from "semantic-ui-react";

import { WorkspacesList_items } from "./__generated__/WorkspacesList_items.graphql";

import WorkspacesListItem from "./WorkspacesListItem";


interface IProps {
  items: WorkspacesList_items;
  onClone: (id: string) => any;
}

export class WorkspacesList extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const cards = items.map((item) => (
      <WorkspacesListItem
        key={item.id}
        item={item}
        onClone={this.handleClone.bind(this, item.id)}
      />
     ));

    return (
      <Card.Group itemsPerRow={3}>
        {cards}
      </Card.Group>
    );
  }

  private handleClone(id: string) {
    this.props.onClone(id);
  }
}

export default createFragmentContainer(WorkspacesList, graphql`
  fragment WorkspacesList_items on Workspace
    @relay(plural: true) {
    ...WorkspacesListItem_item
    id
  }`,
);
