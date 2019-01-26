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
import { Table } from "semantic-ui-react";

import { ProcessGroupTableRowGroup_item } from "./__generated__/ProcessGroupTableRowGroup_item.graphql";

import ProcessGroupTableCells from "./ProcessGroupTableCells";
import ProcessGroupTableProcessCells from "./ProcessGroupTableProcessCells";

interface IProps {
  item: ProcessGroupTableRowGroup_item;
}

export class ProcessGroupTableRowGroup extends Component<IProps> {

  public render() {
    const item = this.props.item;
    const processes = item.processes;
    const groupCells = <ProcessGroupTableCells item={item} />;

    const rows = processes.map((process, index) => {
      const processCells = <ProcessGroupTableProcessCells item={process} />;

      if (index === 0) {
        return (
          <Table.Row key={process.id}>
            {groupCells}
            {processCells}
          </Table.Row>
        );
      }

      return (
        <Table.Row key={process.id}>
          {processCells}
        </Table.Row>
      );
    });

    return <Fragment>{rows}</Fragment>;
  }

}

export default createFragmentContainer(ProcessGroupTableRowGroup, graphql`
  fragment ProcessGroupTableRowGroup_item on ProcessGroup {
    ...ProcessGroupTableCells_item
    processes {
      id
      ...ProcessGroupTableProcessCells_item
    }
  }`,
);
