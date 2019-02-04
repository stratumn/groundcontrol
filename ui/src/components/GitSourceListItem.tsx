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
import { List } from "semantic-ui-react";

import { createFragmentContainer } from "react-relay";

import { GitSourceListItem_item } from "./__generated__/GitSourceListItem_item.graphql";

interface IProps {
  item: GitSourceListItem_item;
}

export class GitSourceListItem extends Component<IProps> {

  public render() {
    const { repository, branch } = this.props.item;

    return (
      <List.Item>
        <List.Content>
          <List.Header>{repository}@{branch}</List.Header>
          <List.Description>Git Repository</List.Description>
        </List.Content>
      </List.Item>
    );
  }

}

export default createFragmentContainer(GitSourceListItem, graphql`
  fragment GitSourceListItem_item on GitSource {
    repository
    branch
  }`,
);
