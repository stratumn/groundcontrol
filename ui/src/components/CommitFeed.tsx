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
import { Feed } from "semantic-ui-react";

import { CommitFeed_items } from "./__generated__/CommitFeed_items.graphql";

import CommitFeedEvent from "./CommitFeedEvent";

interface IProps {
  items: CommitFeed_items;
}

export class CommitFeed extends Component<IProps> {

  public render() {
    const items = this.props.items;

    const rows = items.map((item) => (
      <CommitFeedEvent
        key={item.id}
        item={item}
      />
    ));

    return <Feed>{rows}</Feed>;
  }

}

export default createFragmentContainer(CommitFeed, graphql`
  fragment CommitFeed_items on Commit
    @relay(plural: true) {
    ...CommitFeedEvent_item
    id
  }`,
);
