import graphql from "babel-plugin-relay/macro";
import { Link } from "found";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import {
  Button,
  Card,
  Divider,
  Header,
 } from "semantic-ui-react";

import { WorkspaceCard_item } from "./__generated__/WorkspaceCard_item.graphql";

import ProjectList from "./ProjectList";

interface IProps {
  item: WorkspaceCard_item;
  onClone: () => any;
}

export class WorkspaceCard extends Component<IProps> {

  public render() {
    const item = this.props.item;

    return (
      <Card>
        <Card.Content>
          <Link
            to={`/workspaces/${item.slug}`}
            Component={Card.Header}
          >
            {item.name}
          </Link>
          <Card.Meta>{item.description}</Card.Meta>
          <Divider horizontal={true}>
            <Header as="h6">Repositories</Header>
          </Divider>
          <Card.Description style={{ marginTop: "1em" }}>
            <ProjectList items={item.projects} />
          </Card.Description>
        </Card.Content>
        <Card.Content extra={true}>
          <div className="ui three buttons">
            <Link
              to={`/workspaces/${item.slug}`}
              className="ui grey button"
            >
              Details
            </Link>
            <Button
              color="teal"
              onClick={this.props.onClone}
            >
              Clone
            </Button>
          </div>
        </Card.Content>
      </Card>
    );
  }

}

export default createFragmentContainer(WorkspaceCard, graphql`
  fragment WorkspaceCard_item on Workspace {
    slug
    name
    description
    projects {
      ...ProjectList_items
    }
  }`,
);
