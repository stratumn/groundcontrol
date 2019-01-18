import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import ReactMarkdown from "react-markdown";
import { createFragmentContainer, RelayProp } from "react-relay";
import { Disposable } from "relay-runtime";
import { Label } from "semantic-ui-react";

import { WorkspaceViewPage_viewer } from "./__generated__/WorkspaceViewPage_viewer.graphql";

import Page from "../components/Page";
import ProjectCardGroup from "../components/ProjectCardGroup";
import WorkspaceMenu from "../components/WorkspaceMenu";
import { commit as cloneProject } from "../mutations/cloneProject";
import { commit as cloneWorkspace } from "../mutations/cloneWorkspace";
import { subscribe } from "../subscriptions/workspaceUpdated";

interface IProps {
  relay: RelayProp;
  viewer: WorkspaceViewPage_viewer;
}

export class WorkspaceViewPage extends Component<IProps> {

  private disposables: Disposable[] = [];

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
        <WorkspaceMenu
          workspace={workspace}
          onClone={this.handleCloneWorkspace}
        />
        <ProjectCardGroup
          items={items}
          onClone={this.handleCloneProject}
        />
      </Page>
    );
  }

  public componentDidMount() {
    this.disposables.push(subscribe(this.props.relay.environment, this.props.viewer.workspace!.id));
  }

  public componentWillUnmount() {
    for (const disposable of this.disposables) {
      disposable.dispose();
    }

    this.disposables = [];
  }

  private handleCloneWorkspace = () => {
    cloneWorkspace(this.props.relay.environment, this.props.viewer.workspace!.id);
  }

  private handleCloneProject = (id: string) => {
    cloneProject(this.props.relay.environment, id);
  }
}

export default createFragmentContainer(WorkspaceViewPage, graphql`
  fragment WorkspaceViewPage_viewer on User
    @argumentDefinitions(
      slug: { type: "String!" },
      commitsLimit: { type: "Int", defaultValue: 3 },
    ) {
    workspace(slug: $slug) {
      id
      name
      description
      notes
      ...WorkspaceMenu_workspace
      projects {
        ...ProjectCardGroup_items @arguments(commitsLimit: $commitsLimit)
      }
    }
  }`,
);
