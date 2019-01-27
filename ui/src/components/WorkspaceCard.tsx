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
import { Link } from "found";
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";
import {
  Button,
  Card,
  Divider,
  Header,
  Label,
 } from "semantic-ui-react";

import { WorkspaceCard_item } from "./__generated__/WorkspaceCard_item.graphql";

import ProjectList from "./ProjectList";

import "./WorkspaceCard.css";

interface IProps {
  item: WorkspaceCard_item;
  onClone: () => any;
  onPull: () => any;
}

export class WorkspaceCard extends Component<IProps> {

  public render() {
    const item = this.props.item;
    const labels: JSX.Element[] = [];
    const buttons: JSX.Element[] = [];

    let color: "grey" | "teal" = "grey";

    if (item.isCloned) {
      color = "teal";

      labels.push((
        <Label
          key="cloned"
          content="cloned"
          color="blue"
          size="small"
        />
      ));

      buttons.push((
        <Button
          key="pull"
          content="Pull"
          color="teal"
          icon="download"
          disabled={item.isPulling}
          loading={item.isPulling}
          onClick={this.props.onPull}
        />
      ));
    } else {
      buttons.push((
        <Button
          key="clone"
          content="Clone"
          color="teal"
          icon="clone"
          disabled={item.isCloning}
          loading={item.isCloning}
          onClick={this.props.onClone}
        />
      ));
    }

    return (
      <Card
        className="WorkspaceCard"
        color={color}
      >
        <Card.Content>
          <Link
            to={`/workspaces/${item.slug}`}
            Component={Card.Header}
          >
            {item.name}
          </Link>
          {labels}
          <Card.Description>{item.description}</Card.Description>
          <Divider horizontal={true}>
            <Header as="h6">Repositories</Header>
          </Divider>
          <Card.Description>
            <ProjectList items={item.projects} />
          </Card.Description>
        </Card.Content>
        <Card.Content extra={true}>
          <div className="ui three buttons">
            <Link
              to={`/workspaces/${item.slug}`}
              className="ui grey button"
            >
              Details
            </Link>
            {buttons}
          </div>
        </Card.Content>
      </Card>
    );
  }

}

export default createFragmentContainer(WorkspaceCard, graphql`
  fragment WorkspaceCard_item on Workspace {
    id
    slug
    name
    description
    isCloned
    isCloning
    isPulling
    projects {
      ...ProjectList_items
    }
  }`,
);
