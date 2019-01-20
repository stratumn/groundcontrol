import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import {
  Dropdown,
  Icon,
  Menu,
} from "semantic-ui-react";

import { WorkspaceMenu_workspace } from "./__generated__/WorkspaceMenu_workspace.graphql";

interface IProps {
  workspace: WorkspaceMenu_workspace;
  onClone: () => any;
}

export class WorkspaceMenu extends Component<IProps> {

  public render() {
    const { isCloning, isCloned } = this.props.workspace;

    return (
      <Menu secondary={true}>
        <Menu.Item
          disabled={isCloning || isCloned}
          onClick={this.props.onClone}
        >
          <Icon name="clone" />
          Clone All
        </Menu.Item>
        <Menu.Item disabled={true}>
          <Icon name="download" />
          Pull Outdated
        </Menu.Item>
        <Menu.Item disabled={true}>
          <Icon name="power" />
          Power On
        </Menu.Item>
        <Menu.Item>
          <Dropdown item={true} text="Tasks" pointing={true} disabled={true}>
            <Dropdown.Menu>
              <Dropdown.Item>Run Tests</Dropdown.Item>
              <Dropdown.Item>Clear Database</Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
        </Menu.Item>
      </Menu>
    );
  }

}

export default createFragmentContainer(WorkspaceMenu, graphql`
  fragment WorkspaceMenu_workspace on Workspace {
    isCloning
    isCloned
  }`,
);
