import graphql from "babel-plugin-relay/macro";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import {
  Button,
  Card,
  Divider,
  Header,
  Label,
 } from "semantic-ui-react";

import { ProjectCard_item } from "./__generated__/ProjectCard_item.graphql";

import CommitList from "./CommitList";
import RepositoryShortName from "./RepositoryShortName";

interface IProps {
  item: ProjectCard_item;
}

export class ProjectCard extends Component<IProps> {

  public render() {
    const item = this.props.item;

    const commits = item.commits.edges.map(({ node }) => node);

    return (
      <Card>
        <Card.Content>
          <Card.Header>
            <RepositoryShortName repository={item.repository} />
          </Card.Header>
          <Label style={{ marginTop: ".8em" }}>{item.branch}</Label>
          <Card.Description style={{ marginTop: "1em" }}>
            {item.description || "No description."}
          </Card.Description>
          <Divider horizontal={true}>
            <Header as="h6">Latest Commits</Header>
          </Divider>
          <CommitList items={commits} />
        </Card.Content>
        <Card.Content extra={true}>
          <div className="ui three buttons">
            <Button
              color="teal"
              disabled={true}
            >
              Pull
            </Button>
          </div>
        </Card.Content>
      </Card>
    );
  }

}

export default createFragmentContainer(ProjectCard, graphql`
  fragment ProjectCard_item on Project
    @argumentDefinitions(
      commitsLimit: { type: "Int", defaultValue: 3 },
    ) {
    repository
    branch
    description
    commits(first: $commitsLimit) {
      edges {
        node {
          ...CommitList_items
        }
      }
    }
  }`,
);
