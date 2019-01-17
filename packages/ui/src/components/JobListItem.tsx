import graphql from "babel-plugin-relay/macro";
import { Link } from "found";
import React, { Component } from "react";
import {
  Table,
 } from "semantic-ui-react";

import { createFragmentContainer } from "react-relay";
import { JobListItem_item } from "./__generated__/JobListItem_item.graphql";

import Moment from "react-moment";
import RepositoryShortName from "./RepositoryShortName";

interface IProps {
  item: JobListItem_item;
}

export class JobListItem extends Component<IProps> {

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
        <Table.Cell>
          <RepositoryShortName repository={item.project.repository} />
         </Table.Cell>
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

export default createFragmentContainer(JobListItem, graphql`
  fragment JobListItem_item on Job {
    name
    status
    createdAt
    updatedAt
    project {
      repository
      branch
      workspace {
        slug
        name
      }
    }
  }`,
);
