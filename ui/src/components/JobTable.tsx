// Copyright 2019 Stratumn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import {
  Segment,
  Table,
 } from "semantic-ui-react";

import { JobTable_items } from "./__generated__/JobTable_items.graphql";

import JobTableRow from "./JobTableRow";

interface IProps {
  items: JobTable_items;
  onStop: (id: string) => any;
}

export class JobTable extends Component<IProps> {

  public render() {
    const items = this.props.items;

    if (items.length < 1) {
      return <Segment>There are no jobs at this time.</Segment>;
    }

    const rows = items.map((item) => (
      <JobTableRow
        key={item.id}
        item={item}
        onStop={this.handleStop.bind(this, item.id)}
      />
    ));

    return (
      <Table inverted={true}>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>Name</Table.HeaderCell>
            <Table.HeaderCell>Workspace</Table.HeaderCell>
            <Table.HeaderCell>Repository</Table.HeaderCell>
            <Table.HeaderCell>Branch</Table.HeaderCell>
            <Table.HeaderCell>Created At</Table.HeaderCell>
            <Table.HeaderCell>Updated At</Table.HeaderCell>
            <Table.HeaderCell>Priority</Table.HeaderCell>
            <Table.HeaderCell>Status</Table.HeaderCell>
            <Table.HeaderCell />
          </Table.Row>
        </Table.Header>
        <Table.Body>{rows}</Table.Body>
      </Table>
    );
  }

  private handleStop(id: string) {
    this.props.onStop(id);
  }

}

export default createFragmentContainer(JobTable, graphql`
  fragment JobTable_items on Job
    @relay(plural: true) {
    ...JobTableRow_item
    id
  }`,
);
