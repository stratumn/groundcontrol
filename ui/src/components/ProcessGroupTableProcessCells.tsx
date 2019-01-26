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
import React, { Component, Fragment } from "react";
import { createFragmentContainer } from "react-relay";
import { Button, Table } from "semantic-ui-react";

import { ProcessGroupTableProcessCells_item } from "./__generated__/ProcessGroupTableProcessCells_item.graphql";

interface IProps {
  item: ProcessGroupTableProcessCells_item;
}

export class ProcessGroupTableProcessCells extends Component<IProps> {

  public render() {
    const { command, status } = this.props.item;

    let actions: JSX.Element[] = [];

    switch (status) {
    case "DONE":
    case "FAILED":
      actions.push((
        <Button
          key="start"
          size="tiny"
          color="teal"
        >
          Start Process
        </Button>
      ));
      break;
    case "RUNNING":
      actions.push((
        <Button
          key="stop"
          size="tiny"
          color="pink"
        >
          Stop Process
        </Button>
      ));
      break;
    }

    return (
      <Fragment>
        <Table.Cell>
          {command}
        </Table.Cell>
        <Table.Cell
          positive={status === "DONE"}
          warning={status === "RUNNING"}
          negative={status === "FAILED"}
        >
          {status}
        </Table.Cell>
        <Table.Cell collapsing={true}>
          {actions}
        </Table.Cell>
      </Fragment>
    );
  }

}

export default createFragmentContainer(ProcessGroupTableProcessCells, graphql`
  fragment ProcessGroupTableProcessCells_item on Process {
    command
    status
  }`,
);
