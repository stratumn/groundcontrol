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
import { Loader, SemanticWIDTHS } from "semantic-ui-react";

import { WorkspaceListPage_system } from "./__generated__/WorkspaceListPage_system.graphql";
import { WorkspaceListPage_viewer } from "./__generated__/WorkspaceListPage_viewer.graphql";

import Page from "../components/Page";
import Welcome from "../components/Welcome";
import WorkspaceCardGroup from "../components/WorkspaceCardGroup";
import WorkspaceSearch from "../components/WorkspaceSearch";
import { commit as cloneWorkspace } from "../mutations/cloneWorkspace";
import { commit as pullWorkspace } from "../mutations/pullWorkspace";
import { subscribe as subscribeSourceUpserted } from "../subscriptions/sourceUpserted";
import { subscribe as subscribeWorkspaceUpserted } from "../subscriptions/workspaceUpserted";

interface IProps {
  relay: RelayProp;
  system: WorkspaceListPage_system;
  viewer: WorkspaceListPage_viewer;
}

interface IState {
  query: string;
  itemsPerRow: SemanticWIDTHS;
}

export class WorkspaceListPage extends Component<IProps, IState> {

  public state: IState = {
    itemsPerRow: 3,
    query: "",
  };

  private disposables: Disposable[] = [];

  public render() {
    if (this.props.viewer.sources.edges.length < 1) {
      return <Welcome />;
    }

    const query = this.state.query;
    const itemsPerRow = this.state.itemsPerRow;
    let items = this.props.viewer.workspaces.edges.map(({ node }) => node);
    let isLoading = false;

    for (const { node } of this.props.viewer.sources.edges) {
      if (node.isLoading) {
        isLoading = true;
        break;
      }
    }

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
        <WorkspaceCardGroup
          items={items}
          itemsPerRow={itemsPerRow}
          onClone={this.handleClone}
          onPull={this.handlePull}
        />
        <Loader
          active={isLoading && items.length < 1}
        />
      </Page>
    );
  }

  public componentDidMount() {
    const environment = this.props.relay.environment;
    const lastMessageId = this.props.system.lastMessageId;
    this.disposables.push(subscribeSourceUpserted(environment, lastMessageId));
    this.disposables.push(subscribeWorkspaceUpserted(environment, lastMessageId));

    this.setItemsPerRow();
    window.addEventListener("resize", this.setItemsPerRow);
    this.disposables.push({
      dispose: () => {
        window.removeEventListener("resize", this.setItemsPerRow);
      },
    });
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

  private setItemsPerRow = () => {
    let itemsPerRow = Math.floor(window.innerWidth / 400);
    itemsPerRow = Math.min(Math.max(itemsPerRow, 1), 16);
    this.setState({itemsPerRow: itemsPerRow as SemanticWIDTHS});
  }
}

export const fragments = graphql`
  fragment WorkspaceListPage_source on Source {
    isLoading
  }
`;

export default createFragmentContainer(WorkspaceListPage, graphql`
  fragment WorkspaceListPage_system on System {
    lastMessageId
  }
  fragment WorkspaceListPage_viewer on User {
    sources(first: 10000) {
      edges {
        node {
          ...WorkspaceListPage_source @relay(mask: false)
        }
      }
    }
    workspaces(first: 10000) @connection(key: "WorkspaceListPage_workspaces") {
      edges {
        node {
          ...WorkspaceCardGroup_items
          name
        }
      }
    }
  }`,
);
