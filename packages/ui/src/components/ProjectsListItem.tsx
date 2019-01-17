import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import Moment from "react-moment";
import { createFragmentContainer } from "react-relay";
import {
  Button,
  Card,
  Divider,
  Feed,
  Header,
  Label,
 } from "semantic-ui-react";

import { ProjectsListItem_item } from "./__generated__/ProjectsListItem_item.graphql";

import RepoShortName from "./RepoShortName";

interface IProps {
  item: ProjectsListItem_item;
}

export class ProjectsListItem extends Component<IProps> {

  public render() {
    const item = this.props.item;

    // TODO: move to own components.
    const commits = item.commits.edges.map((edge) => (
      <Feed.Event key={edge.node.id}>
        <Feed.Content>
          <Feed.Summary>
            {edge.node.headline}
          </Feed.Summary>
          <Feed.Meta>
            Pushed by <strong>{edge.node.author}</strong>
            <Moment
              fromNow={true}
              style={{marginLeft: 0}}
            >
              {edge.node.date}
            </Moment>
          </Feed.Meta>
        </Feed.Content>
      </Feed.Event>
    ));

    return (
      <Card>
        <Card.Content>
          <Card.Header>
            <RepoShortName repo={item.repo} />
          </Card.Header>
          <Label style={{ marginTop: ".8em" }}>{item.branch}</Label>
          <Card.Description style={{ marginTop: "1em" }}>
            {item.description || "No description."}
          </Card.Description>
          <Divider horizontal={true}>
            <Header as="h6">Latest Commits</Header>
          </Divider>
          <Feed>
            {commits}
          </Feed>
        </Card.Content>
        <Card.Content extra={true}>
          <div className="ui three buttons">
            <Button color="teal" disabled={true}>
              Pull
            </Button>
          </div>
        </Card.Content>
      </Card>
    );
  }

}

export default createFragmentContainer(ProjectsListItem, graphql`
  fragment ProjectsListItem_item on Project {
    id
    repo
    branch
    description
    commits(first: 3) {
      edges {
        node {
          id
          headline
          date
          author
        }
      }
    }
  }`,
);
