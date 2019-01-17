import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import ReactMarkdown from "react-markdown";
import { createFragmentContainer } from "react-relay";
import {
  Dropdown,
  Header,
  Icon,
  Label,
  Menu,
 } from "semantic-ui-react";

import { WorkspaceViewPage_viewer } from "./__generated__/WorkspaceViewPage_viewer.graphql";

import ProjectCardGroup from "../components/ProjectCardGroup";

interface IProps {
  viewer: WorkspaceViewPage_viewer;
}

export class WorkspaceViewPage extends Component<IProps> {

  public render() {
    const workspace = this.props.viewer.workspace!;
    const items = workspace.projects!;
    const notes = workspace.notes || "No notes";

    // TODO: move to own components.
    return (
      <div>
        <Header as="h1">
          <Icon name="cube" />
          <Header.Content>
            {workspace.name}
            <Header.Subheader>
              {workspace.description}
            </Header.Subheader>
          </Header.Content>
        </Header>
        <Label size="large">not cloned</Label>
        <div style={{ margin: "2em 0" }}>
          <ReactMarkdown source={notes} />
        </div>
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
        <ProjectCardGroup
          items={items}
        />
      </div>
    );
  }

}

export default createFragmentContainer(WorkspaceViewPage, graphql`
  fragment WorkspaceViewPage_viewer on User
    @argumentDefinitions(
      slug: { type: "String!" },
      commitsLimit: { type: "Int", defaultValue: 3 },
    ) {
    workspace(slug: $slug) {
      projects {
        ...ProjectCardGroup_items @arguments(commitsLimit: $commitsLimit)
      }
      name
      description
      notes
    }
  }`,
);
