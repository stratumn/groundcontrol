import graphql from "babel-plugin-relay/macro";
import { Router } from "found";
import React, { Component } from "react";
import {
  Table,
 } from "semantic-ui-react";

import { createFragmentContainer } from "react-relay";

import { JobsList_items } from "./__generated__/JobsList_items.graphql";

import JobsListItem from "./JobsListItem";

interface IProps {
  items: JobsList_items;
}

export class JobsList extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const rows = items.map((item) => <JobsListItem key={item.id} item={item} />);

    return (
      <Table celled={true}>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>Name</Table.HeaderCell>
            <Table.HeaderCell>Created At</Table.HeaderCell>
            <Table.HeaderCell>Updated At</Table.HeaderCell>
            <Table.HeaderCell>Workspace</Table.HeaderCell>
            <Table.HeaderCell>Repo</Table.HeaderCell>
            <Table.HeaderCell>Status</Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {rows}
        </Table.Body>
      </Table>
    );
  }

}

export default createFragmentContainer(JobsList, graphql`
  fragment JobsList_items on Job
    @relay(plural: true) {
    ...JobsListItem_item
    id
  }`,
);
