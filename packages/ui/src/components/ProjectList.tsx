import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import { List } from "semantic-ui-react";

import { ProjectList_items } from "./__generated__/ProjectList_items.graphql";

import ProjectListItem from "./ProjectListItem";

interface IProps {
  items: ProjectList_items;
}

export class ProjectList extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const rows = items.map((item) => (
      <ProjectListItem
        key={item.id}
        item={item}
      />
     ));

    return (
      <List>
        {rows}
      </List>
    );
  }

}

export default createFragmentContainer(ProjectList, graphql`
  fragment ProjectList_items on Project
    @relay(plural: true) {
    ...ProjectListItem_item
    id
  }`,
);
