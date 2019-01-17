import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer, RelayProp } from "react-relay";
import {
  Header,
  Icon,
 } from "semantic-ui-react";

import { WorkspaceListPage_viewer } from "./__generated__/WorkspaceListPage_viewer.graphql";

import Page from "../components/Page";
import WorkspacesCardGroup from "../components/WorkspaceCardGroup";
import WorkspaceSearch from "../components/WorkspaceSearch";
import { commit as cloneWorkspace } from "../mutations/cloneWorkspace";

interface IProps extends WorkspaceListPage {
  relay: RelayProp;
  viewer: WorkspaceListPage_viewer;
}

interface IProps {
  viewer: WorkspaceListPage_viewer;
}

interface IState {
  query: string;
}

export class WorkspaceListPage extends Component<IProps, IState> {

  public state: IState = {
    query: "",
  };

  public render() {
    const query = this.state.query;
    let items = this.props.viewer.workspaces!;

    if (query) {
      items = items.filter((item) => item.name.toLowerCase().indexOf(query) >= 0);
    }

    return (
      <Page
        header="Workspaces"
        subheader="A workspace is a collection of related Github repositories and branches."
        icon="cubes"
      >
        <WorkspaceSearch
          onChange={this.handleSearchChange}
        />
        <WorkspacesCardGroup
          items={items}
          onClone={this.handleClone}
        />
      </Page>
    );
  }

  private handleSearchChange = (query: string) => {
    this.setState({ query });
  }

  private handleClone = (id: string) => {
    cloneWorkspace(this.props.relay.environment, id);
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
