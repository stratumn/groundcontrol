import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import {
  Table,
 } from "semantic-ui-react";

import { JobList_items } from "./__generated__/JobList_items.graphql";

import JobListItem from "./JobListItem";

interface IProps {
  items: JobList_items;
}

export class JobList extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const rows = items.map((item) => (
      <JobListItem
        key={item.id}
        item={item}
      />
    ));

    return (
      <Table celled={true}>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>Name</Table.HeaderCell>
            <Table.HeaderCell>Workspace</Table.HeaderCell>
            <Table.HeaderCell>Repository</Table.HeaderCell>
            <Table.HeaderCell>Branch</Table.HeaderCell>
            <Table.HeaderCell>Created At</Table.HeaderCell>
            <Table.HeaderCell>Updated At</Table.HeaderCell>
            <Table.HeaderCell>Status</Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>{rows}</Table.Body>
      </Table>
    );
  }

}

export default createFragmentContainer(JobList, graphql`
  fragment JobList_items on Job
    @relay(plural: true) {
    ...JobListItem_item
    id
  }`,
);
