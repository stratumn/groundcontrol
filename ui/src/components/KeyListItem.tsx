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
import {
  Button,
  List,
} from "semantic-ui-react";

import { createFragmentContainer } from "react-relay";

import { KeyListItem_item } from "./__generated__/KeyListItem_item.graphql";

import "./KeyListItem.css";

interface IProps {
  item: KeyListItem_item;
  onEdit: () => any;
  onDelete: () => any;
}

export class KeyListItem extends Component<IProps> {

  public render() {
    const { name, value } = this.props.item;
    const { onEdit, onDelete } = this.props;

    return (
      <List.Item className="KeyListItem">
        <List.Content>
          <List.Header>{name}</List.Header>
          <List.Description>
            <code>
              {value}
            </code>
          </List.Description>
          <Button
            icon="edit"
            content="Edit"
            color="teal"
            size="small"
            onClick={onEdit}
          />
          <Button
            icon="delete"
            color="pink"
            size="small"
            onClick={onDelete}
          />
        </List.Content>
      </List.Item>
    );
  }

}

export default createFragmentContainer(KeyListItem, graphql`
  fragment KeyListItem_item on Key {
    name
    value
  }`,
);
