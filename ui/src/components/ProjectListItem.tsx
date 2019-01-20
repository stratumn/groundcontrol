import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import Moment from "react-moment";
import { createFragmentContainer } from "react-relay";
import {
  Label,
  List,
} from "semantic-ui-react";

import { ProjectListItem_item } from "./__generated__/ProjectListItem_item.graphql";

import RepositoryShortName from "./RepositoryShortName";

import "./ProjectListItem.css";

interface IProps {
  item: ProjectListItem_item;
}

export class ProjectListItem extends Component<IProps> {

  public render() {
    const item = this.props.item;

    return (
      <List.Item className="ProjectListItem">
        <List.Content floated="right">
          <Label size="small">{item.branch}</Label>
        </List.Content>
        <List.Content>
          <RepositoryShortName repository={item.repository} />
        </List.Content>
      </List.Item>
    );
  }

}

export default createFragmentContainer(ProjectListItem, graphql`
  fragment ProjectListItem_item on Project {
    repository
    branch
  }`,
);
