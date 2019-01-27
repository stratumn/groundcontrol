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
  Icon,
  Menu,
} from "semantic-ui-react";

import { WorkspaceMenu_workspace } from "./__generated__/WorkspaceMenu_workspace.graphql";

import WorkspaceTaskDropdown from "./WorkspaceTaskDropdown";

interface IProps {
  workspace: WorkspaceMenu_workspace;
  onClone: () => any;
  onPull: () => any;
  onRun: (id: string) => any;
}

export class WorkspaceMenu extends Component<IProps> {

  public render() {
    const { isCloning, isCloned, isPulling, tasks } = this.props.workspace;
    const { onClone, onPull, onRun } = this.props;

    return (
      <Menu size="large">
        <Menu.Item
          disabled={isCloning || isCloned}
          onClick={onClone}
        >
          <Icon name="clone" />
          Clone All
        </Menu.Item>
        <Menu.Item
          disabled={isPulling || !isCloned}
          onClick={onPull}
        >
          <Icon name="download" />
          Pull All
        </Menu.Item>
        <WorkspaceTaskDropdown
          items={tasks}
          enabled={isCloned}
          onRun={onRun}
        />
      </Menu>
    );
  }

}

export default createFragmentContainer(WorkspaceMenu, graphql`
  fragment WorkspaceMenu_workspace on Workspace {
    tasks {
      ...WorkspaceTaskDropdown_items
    }
    isCloning
    isCloned
    isPulling
  }`,
);
