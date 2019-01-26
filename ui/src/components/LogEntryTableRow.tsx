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
import {
  Accordion,
  Icon,
  Table,
 } from "semantic-ui-react";

import Moment from "react-moment";
import { createFragmentContainer } from "react-relay";

import { LogEntryTableRow_item } from "./__generated__/LogEntryTableRow_item.graphql";

import "./LogEntryTableRow.css";

const dateFormat = "L LTS";

interface IProps {
  item: LogEntryTableRow_item;
}

export class LogEntryTableRow extends Component<IProps> {

  public render() {
    const item = this.props.item;
    const panels = [{
      content: JSON.stringify(JSON.parse(item.metaJSON), null, 2),
      key: "details",
      title: item.message,
    }];

    return (
      <Table.Row
        className="LogEntryTableRow"
        verticalAlign="top"
      >
        <Table.Cell className="LogEntryTableRowCreatedAt">
          <Moment format={dateFormat}>{item.createdAt}</Moment>
        </Table.Cell>
        <Table.Cell
          className="LogEntryTableRowLevel"
          warning={item.level === "WARNING"}
          error={item.level === "ERROR"}
        >
          {item.level}
        </Table.Cell>
        <Table.Cell>
          <Accordion panels={panels} />
        </Table.Cell>
      </Table.Row>
    );
  }

}

export default createFragmentContainer(LogEntryTableRow, graphql`
  fragment LogEntryTableRow_item on LogEntry {
    createdAt
    level
    message
    metaJSON
  }`,
);
