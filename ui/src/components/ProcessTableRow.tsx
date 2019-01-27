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
  Button,
  Icon,
  Table,
 } from "semantic-ui-react";

import { ProcessTableRow_item } from "./__generated__/ProcessTableRow_item.graphql";

import RepositoryShortName from "./RepositoryShortName";

import "./ProcessTableRow.css";

interface IProps {
  item: ProcessTableRow_item;
  onStart: () => any;
  onStop: () => any;
}

export class ProcessTableRow extends Component<IProps> {

  public render() {
    const { command, status, project: { repository } } = this.props.item;
    const { onStart, onStop } = this.props;
    const buttons: JSX.Element[] = [];

    switch (status) {
    case "DONE":
    case "FAILED":
      buttons.push((
        <Button
          key="start"
          size="tiny"
          icon={true}
          labelPosition="left"
          onClick={onStart}
        >
          <Icon name="play" />
          Start
        </Button>
      ));
      break;
    case "RUNNING":
      buttons.push((
        <Button
          key="stop"
          size="tiny"
          icon={true}
          labelPosition="left"
          onClick={onStop}
        >
          <Icon name="pause" />
          Stop
        </Button>
      ));
      break;
    }

    return (
      <Table.Row className="ProcessTableRow">
        <Table.Cell collapsing={true}>
          <RepositoryShortName repository={repository} />
        </Table.Cell>
        <Table.Cell
          className="ProcessTableRowCommand"
        >
          {command}
        </Table.Cell>
        <Table.Cell
          positive={status === "DONE"}
          warning={status === "RUNNING"}
          error={status === "FAILED"}
          collapsing={true}
        >
          {status}
        </Table.Cell>
        <Table.Cell collapsing={true}>
          {buttons}
        </Table.Cell>
      </Table.Row>
    );
  }

}

export default createFragmentContainer(ProcessTableRow, graphql`
  fragment ProcessTableRow_item on Process {
    command
    status
    project {
      repository
    }
  }`,
);
