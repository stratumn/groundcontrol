import React, { Component } from "react";
import {
  Dropdown,
  Icon,
  Menu,
} from "semantic-ui-react";

export default class WorkspaceMenu extends Component {

  public render() {
    return (
      <Menu secondary={true}>
        <Menu.Item>
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
