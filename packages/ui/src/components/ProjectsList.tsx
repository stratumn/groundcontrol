import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import { Card } from "semantic-ui-react";

import { ProjectsList_items } from "./__generated__/ProjectsList_items.graphql";

import ProjectsListItem from "./ProjectsListItem";

interface IProps {
  items: ProjectsList_items;
}

export class ProjectsList extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const cards = items.map((item) => (
      <ProjectsListItem
        key={item.id}
        item={item}
      />
     ));

    return (
      <Card.Group itemsPerRow={3}>
        {cards}
      </Card.Group>
    );
  }
}

export default createFragmentContainer(ProjectsList, graphql`
  fragment ProjectsList_items on Project
    @argumentDefinitions(
      commitsLimit: { type: "Int", defaultValue: 3 },
    )
    @relay(plural: true) {
    ...ProjectsListItem_item @arguments(commitsLimit: $commitsLimit)
    id
  }`,
);
