import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import {
  Table,
 } from "semantic-ui-react";

import { JobTable_items } from "./__generated__/JobTable_items.graphql";

import JobTableRow from "./JobTableRow";

interface IProps {
  items: JobTable_items;
}

export class JobTable extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const rows = items.map((item) => (
      <JobTableRow
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

export default createFragmentContainer(JobTable, graphql`
  fragment JobTable_items on Job
    @relay(plural: true) {
    ...JobTableRow_item
    id
  }`,
);
