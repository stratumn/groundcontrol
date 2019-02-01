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
import { createFragmentContainer, RelayProp } from "react-relay";
import { Disposable } from "relay-runtime";

import { WorkspaceListPage_viewer } from "./__generated__/WorkspaceListPage_viewer.graphql";

import Page from "../components/Page";
import WorkspacesCardGroup from "../components/WorkspaceCardGroup";
import WorkspaceSearch from "../components/WorkspaceSearch";
import { commit as cloneWorkspace } from "../mutations/cloneWorkspace";
import { commit as pullWorkspace } from "../mutations/pullWorkspace";
import { subscribe } from "../subscriptions/workspaceUpdated";

interface IProps {
  relay: RelayProp;
  viewer: WorkspaceListPage_viewer;
}

interface IState {
  query: string;
}

export class WorkspaceListPage extends Component<IProps, IState> {

  public state: IState = {
    query: "",
  };

  private disposables: Disposable[] = [];

  public render() {
    const query = this.state.query;
    let items = this.props.viewer.workspaces;

    if (query) {
      items = items.filter((item) => item.name.toLowerCase().indexOf(query) >= 0);
    }

    return (
      <Page
        header="Workspaces"
        subheader="A workspace is a collection of related Git repositories and branches."
        icon="cubes"
      >
        <WorkspaceSearch onChange={this.handleSearchChange} />
        <WorkspacesCardGroup
          items={items}
          onClone={this.handleClone}
          onPull={this.handlePull}
        />
      </Page>
    );
  }

  public componentDidMount() {
    this.disposables.push(subscribe(this.props.relay.environment));
  }

  public componentWillUnmount() {
    for (const disposable of this.disposables) {
      disposable.dispose();
    }

    this.disposables = [];
  }

  private handleSearchChange = (query: string) => {
    this.setState({ query });
  }

  private handleClone = (id: string) => {
    cloneWorkspace(this.props.relay.environment, id);
  }

  private handlePull = (id: string) => {
    pullWorkspace(this.props.relay.environment, id);
  }
}

export default createFragmentContainer(WorkspaceListPage, graphql`
  fragment WorkspaceListPage_viewer on User {
    workspaces {
      ...WorkspaceCardGroup_items
      name
    }
  }`,
);
