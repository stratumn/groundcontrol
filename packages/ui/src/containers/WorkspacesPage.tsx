import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer, RelayProp } from "react-relay";
import {
  Header,
  Icon,
 } from "semantic-ui-react";

import { WorkspacesPage_viewer } from "./__generated__/WorkspacesPage_viewer.graphql";

import WorkspacesList from "../components/WorkspacesList";
import WorkspacesListSearch from "../components/WorkspacesListSearch";
import { commit as cloneWorkspace } from "../mutations/cloneWorkspace";

interface IProps extends WorkspacesPage {
  relay: RelayProp;
  viewer: WorkspacesPage_viewer;
}

interface IProps {
  viewer: WorkspacesPage_viewer;
}

interface IState {
  query: string;
}

export class WorkspacesPage extends Component<IProps, IState> {

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
        <WorkspacesListSearch
          onChange={this.handleSearchChange}
        />
        <WorkspacesList
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

export default createFragmentContainer(WorkspacesPage, graphql`
  fragment WorkspacesPage_viewer on User {
    workspaces {
      ...WorkspacesList_items
      name
    }
  }`,
);
