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
import { Dropdown } from "semantic-ui-react";

import { WorkspaceTaskDropdown_items } from "./__generated__/WorkspaceTaskDropdown_items.graphql";

interface IProps {
  items: WorkspaceTaskDropdown_items;
  enabled: boolean;
  onRun: (id: string) => any;
}

export class WorkspaceTaskDropdown extends Component<IProps> {

  public render() {
    const { enabled, items } = this.props;

    const dropdownItems = items.map((item) => (
      <Dropdown.Item
        onClick={this.handleRun.bind(this, item.id)}
      >
        {item.name}
      </Dropdown.Item>
    ));

    return (
      <Dropdown item={true} text="Tasks" pointing={true} disabled={!enabled}>
        <Dropdown.Menu>
          {dropdownItems}
        </Dropdown.Menu>
      </Dropdown>
    );
  }

  private handleRun(id: string) {
    this.props.onRun(id);
  }

}

export default createFragmentContainer(WorkspaceTaskDropdown, graphql`
  fragment WorkspaceTaskDropdown_items on Task @relay(plural: true) {
    id
    name
  }`,
);
