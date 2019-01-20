import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import { Feed } from "semantic-ui-react";

import { CommitList_items } from "./__generated__/CommitList_items.graphql";

import CommitListItem from "./CommitListItem";

interface IProps {
  items: CommitList_items;
}

export class CommitList extends Component<IProps> {

  public render() {
    const items = this.props.items;

    const rows = items.map((item) => (
      <CommitListItem
        key={item.id}
        item={item}
      />
    ));

    return <Feed>{rows}</Feed>;
  }

}

export default createFragmentContainer(CommitList, graphql`
  fragment CommitList_items on Commit
    @relay(plural: true) {
    ...CommitListItem_item
    id
  }`,
);
