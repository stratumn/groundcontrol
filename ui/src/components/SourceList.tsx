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
import { List } from "semantic-ui-react";

import { SourceList_items } from "./__generated__/SourceList_items.graphql";

import DirectorySourceListItem from "./DirectorySourceListItem";
import GitSourceListItem from "./GitSourceListItem";

interface IProps {
  items: SourceList_items;
  onDelete: (id: string) => any;
}

export class SourceList extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const onDelete = this.props.onDelete;

    if (items.length < 1) {
      return <p>There are no sources at this time.</p>;
    }

    const listItems = items.map((item) => {
      switch (item.__typename) {
      case "DirectorySource":
        return (
          <DirectorySourceListItem
            key={item.id}
            item={item}
            onDelete={onDelete.bind(null, item.id)}
          />
        );
      case "GitSource":
        return (
          <GitSourceListItem
            key={item.id}
            item={item}
            onDelete={onDelete.bind(null, item.id)}
          />
        );
      }
      return null;
    });

    return (
      <List divided={true}>
        {listItems}
      </List>
    );
  }

}

export default createFragmentContainer(SourceList, graphql`
  fragment SourceList_items on Source
    @relay(plural: true) {
    __typename
    id
    ... on DirectorySource {
      ...DirectorySourceListItem_item
    }
    ... on GitSource {
      ...GitSourceListItem_item
    }
  }`,
);
