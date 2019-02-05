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
import { createFragmentContainer } from "react-relay";
import { Segment } from "semantic-ui-react";

import { SourceListPage_viewer } from "./__generated__/SourceListPage_viewer.graphql";

import AddSourceForm from "../components/AddSourceForm";
import Page from "../components/Page";
import SourceList from "../components/SourceList";

interface IProps {
  viewer: SourceListPage_viewer;
}

export class SourceListPage extends Component<IProps> {

  public render() {
    const items = this.props.viewer.sources.edges.map(({ node }) => node);

    return (
      <Page
        header="Sources"
        subheader="A source is a collection of workspaces. It can either be a directory or a Git repository"
        icon="folder open"
      >
        <Segment>
          <h3>Add a New Source</h3>
          <AddSourceForm
            onAddDirectorySource={this.handleAddDirectorySource}
            onAddGitSource={this.handleAddGitSource}
          />
        </Segment>
        <Segment>
          <h3>Current Sources</h3>
          <SourceList items={items} />
        </Segment>
      </Page>
    );
  }

  private handleAddDirectorySource(directory: string) {
    console.log(directory);
  }

  private handleAddGitSource(repository: string, branch: string) {
    console.log(repository, branch);
  }

}

export default createFragmentContainer(SourceListPage, graphql`
  fragment SourceListPage_viewer on User {
    sources {
      edges {
        node {
          ...SourceList_items
        }
      }
    }
  }`,
);
