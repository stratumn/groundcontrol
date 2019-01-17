import graphql from "babel-plugin-relay/macro";
import { Link } from "found";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import {
  Button,
  Card,
  Divider,
  Header,
  Label,
  List,
 } from "semantic-ui-react";

import { WorkspacesListItem_item } from "./__generated__/WorkspacesListItem_item.graphql";

import RepositoryShortName from "./RepositoryShortName";

interface IProps {
  item: WorkspacesListItem_item;
  onClone: () => any;
}

export class WorkspacesListItem extends Component<IProps> {

  public render() {
    const item = this.props.item;

    // TODO: move to own components.
    const projects = item.projects.map((project) => (
      <List.Item key={project.id}>
        <List.Content floated="right">
          <Label
            style={{ position: "relative", top: "-.3em" }}
            size="small"
          >
            {project.branch}
          </Label>
        </List.Content>
        <List.Content>
          <RepositoryShortName repository={project.repository} />
        </List.Content>
      </List.Item>
    ));

    return (
      <Card>
        <Card.Content>
          <Link
            to={`/workspaces/${item.slug}`}
            Component={Card.Header}
          >
            {item.name}
          </Link>
          <Card.Meta>
            {item.description}
          </Card.Meta>
          <Divider horizontal={true}>
            <Header as="h6">Repositories</Header>
          </Divider>
          <Card.Description style={{ marginTop: "1em" }}>
            <List>
              {projects}
            </List>
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

export default createFragmentContainer(WorkspacesListItem, graphql`
  fragment WorkspacesListItem_item on Workspace {
    slug
    name
    description
    projects {
      id
      repository
      branch
    }
  }`,
);
