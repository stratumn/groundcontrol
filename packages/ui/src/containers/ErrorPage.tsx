import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer, RelayProp } from "react-relay";
import { Disposable } from "relay-runtime";

import { WorkspaceListPage_viewer } from "./__generated__/WorkspaceListPage_viewer.graphql";

import Page from "../components/Page";
import WorkspacesCardGroup from "../components/WorkspaceCardGroup";
import WorkspaceSearch from "../components/WorkspaceSearch";
import { commit as cloneWorkspace } from "../mutations/cloneWorkspace";
import { subscribe } from "../subscriptions/workspaceUpdated";

interface IProps {
  error: Error;
}

export default class ErrorPage extends Component<IProps> {

  public render() {
    const error = this.props.error;

    return (
      <Page
        header="Oops"
        subheader="Looks like something's wrong."
        icon="warning"
      >
        <pre>{error.stack}</pre>
      </Page>
    );
  }

}