import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import { Card } from "semantic-ui-react";

import { ProjectCardGroup_items } from "./__generated__/ProjectCardGroup_items.graphql";

import ProjectCard from "./ProjectCard";

interface IProps {
  items: ProjectCardGroup_items;
  onClone: (id: string) => any;
}

export class ProjectCardGroup extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const cards = items.map((item) => (
      <ProjectCard
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

export default createFragmentContainer(ProjectCardGroup, graphql`
  fragment ProjectCardGroup_items on Project
    @argumentDefinitions(
      commitsLimit: { type: "Int", defaultValue: 3 },
    )
    @relay(plural: true) {
    ...ProjectCard_item @arguments(commitsLimit: $commitsLimit)
    id
  }`,
);
