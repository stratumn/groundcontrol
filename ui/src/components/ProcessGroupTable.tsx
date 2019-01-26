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
  Table,
 } from "semantic-ui-react";

import { ProcessGroupTable_items } from "./__generated__/ProcessGroupTable_items.graphql";

import ProcessGroupTableRowGroup from "./ProcessGroupTableRowGroup";

interface IProps {
  items: ProcessGroupTable_items;
}

export class ProcessGroupTable extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const groups = items.map((item) => (
      <ProcessGroupTableRowGroup
        key={item.id}
        item={item}
      />
    ));

    return (
      <Table
        celled={true}
        striped={true}
        structured={true}
      >
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell rowSpan={2}>
              Task
            </Table.HeaderCell>
            <Table.HeaderCell rowSpan={2}>
              Workspace
            </Table.HeaderCell>
            <Table.HeaderCell rowSpan={2}>
              Created At
            </Table.HeaderCell>
            <Table.HeaderCell rowSpan={2}>
              Status
            </Table.HeaderCell>
            <Table.HeaderCell rowSpan={2}>
              Actions
            </Table.HeaderCell>
            <Table.HeaderCell colSpan={3}>
              Processes
            </Table.HeaderCell>
          </Table.Row>
          <Table.Row>
            <Table.HeaderCell>Command</Table.HeaderCell>
            <Table.HeaderCell>Status</Table.HeaderCell>
            <Table.HeaderCell>Actions</Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>{groups}</Table.Body>
      </Table>
    );
  }

}

export default createFragmentContainer(ProcessGroupTable, graphql`
  fragment ProcessGroupTable_items on ProcessGroup
    @relay(plural: true) {
    ...ProcessGroupTableRowGroup_item
    id
  }`,
);
