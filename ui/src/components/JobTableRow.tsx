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
import {
  Button,
  Icon,
  Table,
 } from "semantic-ui-react";

import { createFragmentContainer } from "react-relay";

import { JobTableRow_item } from "./__generated__/JobTableRow_item.graphql";

import Moment from "react-moment";
import RepositoryShortName from "./RepositoryShortName";

import "./JobTableRow.css";

const dateFormat = "L LTS";

interface IProps {
  item: JobTableRow_item;
  onStop: () => any;
}

export class JobTableRow extends Component<IProps> {

  public render() {
    const { item, onStop } = this.props;
    const owner = item.owner;
    const buttons: JSX.Element[] = [];

    let workspaceSlug = "-";
    let workspaceName = "-";
    let projectRepository = "-";
    let projectBranch = "-";

    switch (owner.__typename) {
    case "Workspace":
      workspaceSlug = owner.slug;
      workspaceName = owner.name;
      break;
    case "Project":
      workspaceSlug = owner.workspace.slug;
      workspaceName = owner.workspace.name;
      projectRepository = owner.repository;
      projectBranch = owner.branch;
      break;
    }

    if (item.status === "RUNNING") {
      buttons.push((
        <Button
          key="stop"
          size="tiny"
          icon={true}
          labelPosition="left"
          onClick={onStop}
        >
          <Icon name="pause" />
          Stop
        </Button>
      ));
    }

    return (
      <Table.Row className="JobTableRow">
        <Table.Cell>{item.name}</Table.Cell>
        <Table.Cell>
          <Link to={`/workspaces/${workspaceSlug}`}>
            {workspaceName}
          </Link>
        </Table.Cell>
        <Table.Cell>
          <RepositoryShortName repository={projectRepository} />
        </Table.Cell>
        <Table.Cell>{projectBranch}</Table.Cell>
        <Table.Cell>
          <Moment format={dateFormat}>{item.createdAt}</Moment>
        </Table.Cell>
        <Table.Cell>
          <Moment format={dateFormat}>{item.updatedAt}</Moment>
        </Table.Cell>
        <Table.Cell>{item.priority}</Table.Cell>
        <Table.Cell
          positive={item.status === "DONE"}
          warning={item.status === "RUNNING"}
          error={item.status === "FAILED"}
        >
          {item.status}
        </Table.Cell>
        <Table.Cell>
          {buttons}
        </Table.Cell>
      </Table.Row>
    );
  }

}

export default createFragmentContainer(JobTableRow, graphql`
  fragment JobTableRow_item on Job {
    name
    createdAt
    updatedAt
    owner {
      __typename
      ... on Workspace {
        slug
        name
      }
      ... on Project {
        repository
        branch
        workspace {
          slug
          name
        }
      }
    }
    priority
    status
  }`,
);
