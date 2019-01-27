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

import { ProcessTable_items } from "./__generated__/ProcessTable_items.graphql";

import ProcessTableRow from "./ProcessTableRow";

import "./ProcessTable.css";

interface IProps {
  items: ProcessTable_items;
  onStart: (id: string) => any;
  onStop: (id: string) => any;
}

export class ProcessTable extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const rows = items.map((item) => (
      <ProcessTableRow
        key={item.id}
        item={item}
        onStart={this.handleStart.bind(this, item.id)}
        onStop={this.handleStop.bind(this, item.id)}
      />
    ));

    return (
      <Table
        className="ProcessTable"
        inverted={true}
      >
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>Project</Table.HeaderCell>
            <Table.HeaderCell>Command</Table.HeaderCell>
            <Table.HeaderCell>Status</Table.HeaderCell>
            <Table.HeaderCell />
          </Table.Row>
        </Table.Header>
        <Table.Body>{rows}</Table.Body>
      </Table>
    );
  }

  private handleStart(id: string) {
    this.props.onStart(id);
  }

  private handleStop(id: string) {
    this.props.onStop(id);
  }

}

export default createFragmentContainer(ProcessTable, graphql`
  fragment ProcessTable_items on Process
    @relay(plural: true) {
    ...ProcessTableRow_item
    id
  }`,
);
