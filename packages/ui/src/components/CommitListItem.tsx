import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import Moment from "react-moment";
import { createFragmentContainer } from "react-relay";
import { Feed } from "semantic-ui-react";

import { CommitListItem_item } from "./__generated__/CommitListItem_item.graphql";

interface IProps {
  item: CommitListItem_item;
}

export class CommitListItem extends Component<IProps> {

  public render() {
    const item = this.props.item;

    return (
      <Feed.Event>
        <Feed.Content>
          <Feed.Summary>
            {item.headline}
          </Feed.Summary>
          <Feed.Meta>
            Pushed by <strong>{item.author}</strong>
            <Moment
              fromNow={true}
              style={{marginLeft: 0}}
            >
              {item.date}
            </Moment>
          </Feed.Meta>
        </Feed.Content>
      </Feed.Event>
    );
  }

}

export default createFragmentContainer(CommitListItem, graphql`
  fragment CommitListItem_item on Commit {
    headline
    date
    author
  }`,
);
