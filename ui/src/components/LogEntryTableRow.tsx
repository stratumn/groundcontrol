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
import { Link } from "found";
import React, { Component } from "react";
import {
  Table,
 } from "semantic-ui-react";

import { createFragmentContainer } from "react-relay";

import { LogEntryTableRow_item } from "./__generated__/LogEntryTableRow_item.graphql";

import Moment from "react-moment";
import RepositoryShortName from "./RepositoryShortName";

const dateFormat = "L LTS";

interface IProps {
  item: LogEntryTableRow_item;
}

export class LogEntryTableRow extends Component<IProps> {

  public render() {
    const item = this.props.item;

    return (
      <Table.Row>
        <Table.Cell>
          <Moment format={dateFormat}>{item.createdAt}</Moment>
        </Table.Cell>
        <Table.Cell
          warning={item.level === "WARNING"}
          negative={item.level === "ERROR"}
        >
          {item.level}
        </Table.Cell>
        <Table.Cell>{item.message}</Table.Cell>
        <Table.Cell>
          <code>
            <pre>
              {item.metaJSON ? JSON.stringify(JSON.parse(item.metaJSON), null, 2) : ""}
            </pre>
          </code>
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
