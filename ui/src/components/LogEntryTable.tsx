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

import { LogEntryTable_items } from "./__generated__/LogEntryTable_items.graphql";

import LogEntryTableRow from "./LogEntryTableRow";

interface IProps {
  items: LogEntryTable_items;
}

export class LogEntryTable extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const rows = items.map((item) => (
      <LogEntryTableRow
        key={item.id}
        item={item}
      />
    ));

    return (
      <Table inverted={true}>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>Created At</Table.HeaderCell>
            <Table.HeaderCell>Level</Table.HeaderCell>
            <Table.HeaderCell>Message</Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>{rows}</Table.Body>
      </Table>
    );
  }

}

export default createFragmentContainer(LogEntryTable, graphql`
  fragment LogEntryTable_items on LogEntry
    @relay(plural: true) {
    ...LogEntryTableRow_item
    id
  }`,
);
