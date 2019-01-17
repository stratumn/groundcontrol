import graphql from "babel-plugin-relay/macro";
import { Link } from "found";
import React, { Component } from "react";
import {
  Table,
 } from "semantic-ui-react";

import { createFragmentContainer } from "react-relay";
import { JobsListItem_item } from "./__generated__/JobsListItem_item.graphql";

import Moment from "react-moment";
import RepoShortName from "./RepoShortName";

interface IProps {
  item: JobsListItem_item;
}

export class JobsListItem extends Component<IProps> {

  public render() {
    const item = this.props.item;

    return (
      <Table.Row>
        <Table.Cell>{item.name}</Table.Cell>
        <Table.Cell><Moment>{item.createdAt}</Moment></Table.Cell>
        <Table.Cell><Moment>{item.updatedAt}</Moment></Table.Cell>
        <Table.Cell>
          <Link to={`/workspaces/${item.project.workspace.slug}`}>{item.project.workspace.name}</Link>
        </Table.Cell>
        <Table.Cell><RepoShortName repo={item.project.repo} /></Table.Cell>
        <Table.Cell
          positive={item.status === "DONE"}
          warning={item.status === "RUNNING"}
          negative={item.status === "FAILED"}
        >
          {item.status}
        </Table.Cell>
      </Table.Row>
    );
  }

}

export default createFragmentContainer(JobsListItem, graphql`
  fragment JobsListItem_item on Job {
    id
    name
    status
    createdAt
    updatedAt
    project {
      repo
      branch
      workspace {
        slug
        name
      }
    }
  }`,
);
