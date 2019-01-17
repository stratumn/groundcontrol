import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer, RelayProp } from "react-relay";
import {
  Header,
  Icon,
 } from "semantic-ui-react";

import { WorkspaceListPage_viewer } from "./__generated__/WorkspaceListPage_viewer.graphql";

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
      <div>
        <Header as="h1" style={{ marginBottom: "1.2em" }} >
          <Icon name="cubes" />
          <Header.Content>
            Workspaces
            <Header.Subheader>
              A workspace is a collection of related projects. Each project is linked to a Github repository and branch.
            </Header.Subheader>
          </Header.Content>
        </Header>
        <WorkspaceSearch
          onChange={this.handleSearchChange}
        />
        <WorkspacesCardGroup
          items={items}
          onClone={this.handleClone}
        />
      </div>
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
