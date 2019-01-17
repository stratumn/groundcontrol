import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import ReactMarkdown from "react-markdown";
import { createFragmentContainer } from "react-relay";
import { Label } from "semantic-ui-react";

import { WorkspaceViewPage_viewer } from "./__generated__/WorkspaceViewPage_viewer.graphql";

import Page from "../components/Page";
import ProjectCardGroup from "../components/ProjectCardGroup";
import WorkspaceMenu from "../components/WorkspaceMenu";

interface IProps {
  viewer: WorkspaceViewPage_viewer;
}

export class WorkspaceViewPage extends Component<IProps> {

  public render() {
    const workspace = this.props.viewer.workspace!;
    const items = workspace.projects!;
    const notes = workspace.notes || "No notes";

    return (
      <Page
        header={workspace.name}
        subheader={workspace.description}
        icon="cube"
      >
        <Label size="large">not cloned</Label>
        <div style={{ margin: "2em 0" }}>
          <ReactMarkdown source={notes} />
        </div>
        <WorkspaceMenu />
        <ProjectCardGroup items={items} />
      </Page>
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
