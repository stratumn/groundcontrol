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
import { Router } from "found";
import React, { Component } from "react";
import { createPaginationContainer, RelayPaginationProp } from "react-relay";
import { Disposable } from "relay-runtime";
import { Button } from "semantic-ui-react";

import { ProcessGroupListPage_system } from "./__generated__/ProcessGroupListPage_system.graphql";

import Page from "../components/Page";
import ProcessGroupCardGroup from "../components/ProcessGroupCardGroup";
import ProcessGroupFilter from "../components/ProcessGroupFilter";
import { commit as startProcess} from "../mutations/startProcess";
import { commit as startProcessGroup } from "../mutations/startProcessGroup";
import { commit as stopProcess} from "../mutations/stopProcess";
import { commit as stopProcessGroup } from "../mutations/stopProcessGroup";
import { subscribe } from "../subscriptions/processGroupUpserted";

interface IProps {
  relay: RelayPaginationProp;
  router: Router;
  system: ProcessGroupListPage_system;
  params: {
    filters: string | undefined;
  };
}

export class ProcessGroupListPage extends Component<IProps> {

  private disposables: Disposable[] = [];

  public render() {
    const items = this.props.system.processGroups.edges.map(({ node }) => node);
    const filters = this.props.params.filters === undefined ? undefined :
      this.props.params.filters.split(",");

    return (
      <Page
        header="Processes"
        subheader="Processes may be launched by tasks and run in the background"
        icon="list"
      >
        <ProcessGroupFilter
          filters={filters}
          onChange={this.handleFiltersChange}
        />
        <ProcessGroupCardGroup
          items={items}
          onStartGroup={this.handleStartProcessGroup}
          onStopGroup={this.handleStopProcessGroup}
          onStartProcess={this.handleStartProcess}
          onStopProcess={this.handleStopProcess}
        />
        <Button
          disabled={!this.props.relay.hasMore() || this.props.relay.isLoading()}
          loading={this.props.relay.isLoading()}
          color="grey"
          onClick={this.handleLoadMore}
        >
          Load More
        </Button>
      </Page>
    );
  }

  public componentDidMount() {
    const environment = this.props.relay.environment;
    const lastMessageId = this.props.system.lastMessageId;
    this.disposables.push(subscribe(environment, lastMessageId));
  }

  public componentWillUnmount() {
    for (const disposable of this.disposables) {
      disposable.dispose();
    }

    this.disposables = [];
  }

  private handleFiltersChange = (filters: string[]) => {
    if (filters.length < 1 || filters.length > 2) {
      return this.props.router.replace("/processes");
    }

    this.props.router.replace(`/processes/${filters.join(",")}`);
  }

  private handleLoadMore = () => {
    this.props.relay.loadMore(
      10,
      (err) => {
        if (err) {
          console.log(err);
        }

        // Make sure load more button updates.
        this.forceUpdate();
      },
    );
  }

  private handleStartProcessGroup = (id: string) => {
    startProcessGroup(this.props.relay.environment, id);
  }

  private handleStopProcessGroup = (id: string) => {
    stopProcessGroup(this.props.relay.environment, id);
  }

  private handleStartProcess = (id: string) => {
    startProcess(this.props.relay.environment, id);
  }

  private handleStopProcess = (id: string) => {
    stopProcess(this.props.relay.environment, id);
  }

}

export default createPaginationContainer(
  ProcessGroupListPage,
  graphql`
    fragment ProcessGroupListPage_system on System
      @argumentDefinitions(
        count: {type: "Int", defaultValue: 10},
        cursor: {type: "String"},
        status: { type: "[ProcessStatus!]", defaultValue: null },
      ) {
      lastMessageId
      processGroups(
       first: $count,
       after: $cursor,
       status: $status,
      )
        @connection(
          key: "ProcessGroupListPage_processGroups",
          filters: ["status"],
        ) {
        edges {
          node {
            ...ProcessGroupCardGroup_items
            id
          }
        }
      }
    }`,
  {
    direction: "forward",
    getConnectionFromProps: (props) => props.system && props.system.processGroups,
    getVariables: (_, {count, cursor}, fragmentVariables) => ({
      count,
      cursor,
      status: fragmentVariables.status,
    }),
    query: graphql`
      query ProcessGroupListPagePaginationQuery(
        $count: Int!,
        $cursor: String,
        $status: [ProcessStatus!],
      ) {
        system {
          ...ProcessGroupListPage_system @arguments(
            count: $count,
            cursor: $cursor,
            status: $status,
          )
        }
      }
    `,
  },
);
