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
import { Segment, SemanticWIDTHS } from "semantic-ui-react";

import { WorkspaceViewPage_system } from "./__generated__/WorkspaceViewPage_system.graphql";
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
  system: WorkspaceViewPage_system;
  viewer: WorkspaceViewPage_viewer;
}

interface IState {
  itemsPerRow: SemanticWIDTHS;
}

export class WorkspaceViewPage extends Component<IProps, IState> {

  public state: IState = {
    itemsPerRow: 3,
  };

  private disposables: Disposable[] = [];

  public render() {
    const workspace = this.props.viewer.workspace!;
    const items = workspace.projects.edges.map(({ node }) => node);
    const itemsPerRow = this.state.itemsPerRow;
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
          itemsPerRow={itemsPerRow}
          onClone={this.handleCloneProject}
          onPull={this.handlePullProject}
        />
      </Page>
    );
  }

  public componentDidMount() {
    const environment = this.props.relay.environment;
    const lastMessageId = this.props.system.lastMessageId;
    const id = this.props.viewer.workspace!.id;
    this.disposables.push(subscribe(environment, lastMessageId, id));

    this.setItemsPerRow();
    window.addEventListener("resize", this.setItemsPerRow);
    this.disposables.push({
      dispose: () => {
        window.removeEventListener("resize", this.setItemsPerRow);
      },
    });

    loadWorkspaceCommits(environment, id);
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
    if (!this.doesTaskHaveVariables(id)) {
      run(this.props.relay.environment, id);
    }

    // TODO: show modal for variables.
  }

  private doesTaskHaveVariables(id: string) {
    const task = this.props.viewer.workspace!.tasks.edges.find((value) => {
      return value.node.id === id;
    });

    if (task) {
      return task.node.variables.edges.length > 0;
    }

    return false;
  }

  private setItemsPerRow = () => {
    let itemsPerRow = Math.floor(window.innerWidth / 400);
    itemsPerRow = Math.min(Math.max(itemsPerRow, 1), 16);
    this.setState({itemsPerRow: itemsPerRow as SemanticWIDTHS});
  }
}

export default createFragmentContainer(WorkspaceViewPage, graphql`
  fragment WorkspaceViewPage_system on System {
    lastMessageId
  }
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
      tasks {
        edges {
          node {
            id
            variables {
              edges {
                node {
                  id
                }
              }
            }
          }
        }
      }
      projects {
        edges {
          node {
            ...ProjectCardGroup_items @arguments(commitsLimit: $commitsLimit)
          }
        }
      }
      ...WorkspaceMenu_workspace
    }
    keys {
      edges {
        node {
          name
        }
      }
    }
  }`,
);
