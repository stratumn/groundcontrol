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
import {
  Button,
  Card,
  Dimmer,
  Divider,
  Header,
  Label,
  Loader,
 } from "semantic-ui-react";

import { ProjectCard_item } from "./__generated__/ProjectCard_item.graphql";

import CommitFeed from "./CommitFeed";
import RepositoryShortName from "./RepositoryShortName";

import "./ProjectCard.css";

interface IProps {
  item: ProjectCard_item;
  onClone: () => any;
}

export class ProjectCard extends Component<IProps> {

  public render() {
    const item = this.props.item;
    const isLoading = item.commits.isLoading;
    const commits = item.commits.edges.map(({ node }) => node);
    const buttons: JSX.Element[] = [];

    if (!item.isCloned) {
      buttons.push((
        <Button
          key="clone"
          content="Clone"
          color="teal"
          disabled={item.isCloning}
          loading={item.isCloning}
          onClick={this.props.onClone}
        />
      ));
    }

    return (
      <Card className="ProjectCard">
        <Dimmer
          active={isLoading}
          inverted={true}
        >
          <Loader />
        </Dimmer>
        <Card.Content>
          <Card.Header>
            <RepositoryShortName repository={item.repository} />
          </Card.Header>
          <Label color="blue">{item.branch}</Label>
          <Card.Description>
            {item.description || "No description."}
          </Card.Description>
          <Divider horizontal={true}>
            <Header as="h6">Latest Commits</Header>
          </Divider>
          <CommitFeed items={commits} />
        </Card.Content>
        <Card.Content extra={true}>
          <div className="ui three buttons">
            {buttons}
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
    id
    repository
    branch
    description
    isCloning
    isCloned
    commits(first: $commitsLimit) {
      edges {
        node {
          ...CommitFeed_items
        }
      }
      isLoading
    }
  }`,
);
