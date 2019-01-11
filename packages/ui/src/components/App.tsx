import { Link } from "found";
import React, { Component } from "react";
import { Container, Label, Menu } from "semantic-ui-react";

import "./App.css";
export default class App extends Component {
  public render() {
    return (
      <div>
        <Menu fixed="top">
          <Container>
            <Link
              to="/workspaces"
              Component={Menu.Item}
              activePropName="active"
            >
              Workspaces
              <Label color="orange">1</Label>
            </Link>
            <Link
              to="/jobs"
              Component={Menu.Item}
              activePropName="active"
            >
              Jobs
              <Label color="teal">3</Label>
            </Link>
            <Link
              to="/processes"
              Component={Menu.Item}
              activePropName="active"
            >
              Processes
              <Label color="teal">1</Label>
            </Link>
            <Link
              to="/errors"
              Component={Menu.Item}
              activePropName="active"
            >
              Logs
              <Label color="red">2</Label>
            </Link>
            <Menu.Item
              as="a"
              href="http://localhost:4000/graphql"
            >
              GraphQL
            </Menu.Item>
          </Container>
        </Menu>
        <Container style={{ marginTop: "7em" }}>
          {this.props.children}
        </Container>
      </div>
    );
  }
}
