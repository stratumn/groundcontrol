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
import ReactMarkdown from "react-markdown";
import { createFragmentContainer, RelayProp } from "react-relay";
import { Disposable } from "relay-runtime";
import { Segment } from "semantic-ui-react";

import { WorkspaceViewPage_viewer } from "./__generated__/WorkspaceViewPage_viewer.graphql";

import Page from "../components/Page";
import ProjectCardGroup from "../components/ProjectCardGroup";
import WorkspaceMenu from "../components/WorkspaceMenu";
import { commit as cloneProject } from "../mutations/cloneProject";
import { commit as cloneWorkspace } from "../mutations/cloneWorkspace";
import { commit as loadWorkspaceCommits } from "../mutations/loadWorkspaceCommits";
import { commit as pullProject } from "../mutations/pullProject";
import { commit as pullWorkspace } from "../mutations/pullWorkspace";
import { commit as run } from "../mutations/run";
import { subscribe } from "../subscriptions/workspaceUpserted";

import "./WorkspaceViewPage.css";

interface IProps {
  relay: RelayProp;
  viewer: WorkspaceViewPage_viewer;
}

export class WorkspaceViewPage extends Component<IProps> {

  private disposables: Disposable[] = [];

  public render() {
    const workspace = this.props.viewer.workspace!;
    const items = workspace.projects.edges.map(({ node }) => node);
    const notes = workspace.notes || "No notes";

    return (
      <Page
        className="WorkspaceViewPage"
        header={workspace.name}
        subheader={workspace.description || "No description."}
        icon="cube"
      >
        <WorkspaceMenu
          workspace={workspace}
          onClone={this.handleCloneWorkspace}
          onPull={this.handlePullWorkspace}
          onRun={this.handleRun}
        />
        <Segment className="WorkspaceViewPageDescription">
          <ReactMarkdown
            source={notes}
            className="markdown-body"
          />
        </Segment>
        <ProjectCardGroup
          items={items}
          onClone={this.handleCloneProject}
          onPull={this.handlePullProject}
        />
      </Page>
    );
  }

  public componentDidMount() {
    this.disposables.push(subscribe(this.props.relay.environment, this.props.viewer.workspace!.id));
    // The timeout is to give subscriptions a chance to connect before something happens.
    setTimeout(
      () => loadWorkspaceCommits(this.props.relay.environment, this.props.viewer.workspace!.id),
      1000,
    );
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

  private handlePullWorkspace = () => {
    pullWorkspace(this.props.relay.environment, this.props.viewer.workspace!.id);
  }

  private handlePullProject = (id: string) => {
    pullProject(this.props.relay.environment, id);
  }

  private handleRun = (id: string) => {
    run(this.props.relay.environment, id);
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
        edges {
          node {
            ...ProjectCardGroup_items @arguments(commitsLimit: $commitsLimit)
          }
        }
      }
    }
  }`,
);
