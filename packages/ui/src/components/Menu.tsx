import { Link } from "found";
import React, { Component } from "react";
import { Container, Label, Menu as SemanticMenu } from "semantic-ui-react";

export default class Menu extends Component {
  public render() {
    return (
      <SemanticMenu fixed="top" inverted={true} color="grey">
        <Container>
          <Link
            to="/workspaces"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Workspaces
            <Label color="orange">1</Label>
          </Link>
          <Link
            to="/jobs"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Jobs
            <Label color="teal">3</Label>
          </Link>
          <Link
            to="/processes"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Processes
            <Label color="teal">1</Label>
          </Link>
          <Link
            to="/errors"
            Component={SemanticMenu.Item}
            activePropName="active"
          >
            Logs
            <Label color="red">2</Label>
          </Link>
          <SemanticMenu.Item
            as="a"
            href="http://localhost:4000/graphql"
          >
            GraphQL
          </SemanticMenu.Item>
        </Container>
      </SemanticMenu>
    );
  }
}
