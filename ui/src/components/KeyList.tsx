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

import { KeyList_items } from "./__generated__/KeyList_items.graphql";

import KeyListItem from "./KeyListItem";

interface IProps {
  items: KeyList_items;
  onEdit: (id: string) => any;
  onDelete: (id: string) => any;
}

export class KeyList extends Component<IProps> {

  public render() {
    const items = this.props.items;
    const { onEdit, onDelete } = this.props;

    if (items.length < 1) {
      return <p>There are no keys at this time.</p>;
    }

    const listItems = items.map((item) => (
      <KeyListItem
        key={item.id}
        item={item}
        onEdit={onEdit.bind(null, item.id)}
        onDelete={onDelete.bind(null, item.id)}
      />
    ));

    return (
      <List divided={true}>
        {listItems}
      </List>
    );
  }

}

export default createFragmentContainer(KeyList, graphql`
  fragment KeyList_items on Key
    @relay(plural: true) {
    __typename
    id
    ...KeyListItem_item
  }`,
);
