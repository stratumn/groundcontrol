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
import Moment from "react-moment";
import { createFragmentContainer } from "react-relay";
import {
  Button,
  Card,
  Icon,
 } from "semantic-ui-react";

import { ProcessGroupCard_item } from "./__generated__/ProcessGroupCard_item.graphql";

import ProcessTable from "./ProcessTable";

import "./ProcessGroupCard.css";

const dateFormat = "L LTS";

interface IProps {
  item: ProcessGroupCard_item;
  onStartGroup: () => any;
  onStopGroup: () => any;
  onStartProcess: (id: string) => any;
  onStopProcess: (id: string) => any;
}

export class ProcessGroupCard extends Component<IProps> {

  public render() {
    const {
      createdAt,
      status,
      task,
      task: { workspace },
    } = this.props.item;
    const processes = this.props.item.processes.edges.map(({ node }) => node);
    const {
      onStartGroup,
      onStopGroup,
      onStartProcess,
      onStopProcess,
    } = this.props;
    const buttons: JSX.Element[] = [];

    switch (status) {
    case "RUNNING":
      buttons.push((
        <Button
          key="stop"
          color="teal"
          floated="right"
          icon={true}
          labelPosition="left"
          onClick={onStopGroup}
        >
          <Icon name="pause" />
          Stop Group
        </Button>
      ));
      break;
    case "DONE":
    case "FAILED":
      buttons.push((
        <Button
          key="start"
          color="teal"
          floated="right"
          icon={true}
          labelPosition="left"
          onClick={onStartGroup}
        >
          <Icon name="play" />
          Start Group
        </Button>
      ));
      break;
    }

    return (
      <Card
        className="ProcessGroupCard"
        fluid={true}
      >
        <Card.Content>
          <Card.Header>
            {buttons}
            <Link to={`/workspaces/${workspace.slug}`}>
              {workspace.name}
            </Link> / {task.name}
          </Card.Header>
          <Card.Meta>
            <Moment format={dateFormat}>{createdAt}</Moment>
          </Card.Meta>
        </Card.Content>
        <ProcessTable
          items={processes}
          onStart={onStartProcess}
          onStop={onStopProcess}
        />
      </Card>
    );
  }

}

export default createFragmentContainer(ProcessGroupCard, graphql`
  fragment ProcessGroupCard_item on ProcessGroup {
    createdAt
    status
    task {
      name
      workspace {
        slug
        name
      }
    }
    processes {
      edges {
        node {
          ...ProcessTable_items
        }
      }
    }
  }`,
);
