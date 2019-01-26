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
import React, { Component, Fragment } from "react";
import Moment from "react-moment";
import { createFragmentContainer } from "react-relay";
import { Button, Table } from "semantic-ui-react";

import { ProcessGroupTableCells_item } from "./__generated__/ProcessGroupTableCells_item.graphql";

const dateFormat = "L LTS";

interface IProps {
  item: ProcessGroupTableCells_item;
}

export class ProcessGroupTableCells extends Component<IProps> {

  public render() {
    const item = this.props.item;
    const task = item.task;
    const workspace = task.workspace;
    const processesLength = item.processes.length;

    let actions: JSX.Element[] = [];

    switch (item.status) {
    case "DONE":
    case "FAILED":
      actions.push((
        <Button
          key="start"
          size="tiny"
          color="teal"
        >
          Start Group
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
          Stop Group
        </Button>
      ));
      break;
    }

    return (
      <Fragment>
        <Table.Cell rowSpan={processesLength}>
          {task.name}
        </Table.Cell>
        <Table.Cell rowSpan={processesLength}>
          <Link to={`/workspaces/${workspace.slug}`}>
            {workspace.name}
          </Link>
        </Table.Cell>
        <Table.Cell rowSpan={processesLength}>
          <Moment format={dateFormat}>{item.createdAt}</Moment>
        </Table.Cell>
        <Table.Cell
          rowSpan={processesLength}
          positive={item.status === "DONE"}
          warning={item.status === "RUNNING"}
          negative={item.status === "FAILED"}
        >
          {item.status}
        </Table.Cell>
        <Table.Cell
          rowSpan={processesLength}
          collapsing={true}
        >
          {actions}
        </Table.Cell>
      </Fragment>
    );
  }

}

export default createFragmentContainer(ProcessGroupTableCells, graphql`
  fragment ProcessGroupTableCells_item on ProcessGroup {
    createdAt
    status
    task {
      name
      workspace {
        slug
        name
      }
    }
    processes {
      id
    }
  }`,
);
