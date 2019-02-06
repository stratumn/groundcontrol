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

import { JobListPage_system } from "./__generated__/JobListPage_system.graphql";

import JobFilter from "../components/JobFilter";
import JobTable from "../components/JobTable";
import Page from "../components/Page";
import { commit as stopJob } from "../mutations/stopJob";

import { subscribe } from "../subscriptions/jobUpserted";

interface IProps {
  relay: RelayPaginationProp;
  router: Router;
  system: JobListPage_system;
  params: {
    filters: string | undefined;
  };
}

export class JobListPage extends Component<IProps> {

  private disposables: Disposable[] = [];

  public render() {
    const items = this.props.system.jobs.edges.map(({ node }) => node);
    const filters = this.props.params.filters === undefined ? undefined :
      this.props.params.filters.split(",");

    return (
      <Page
        header="Jobs"
        subheader="Jobs are short lived tasks such as cloning a repository."
        icon="tasks"
      >
        <JobFilter
          filters={filters}
          onChange={this.handleFiltersChange}
        />
        <JobTable
          items={items}
          onStop={this.handleStop}
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
    if (filters.length < 1 || filters.length > 4) {
      return this.props.router.replace("/jobs");
    }

    this.props.router.replace(`/jobs/${filters.join(",")}`);
  }

  private handleStop = (id: string) => {
    stopJob(this.props.relay.environment, id);
  }

  private handleLoadMore = () => {
    this.props.relay.loadMore(
      20,
      (err) => {
        if (err) {
          console.log(err);
        }

        // Make sure load more button updates.
        this.forceUpdate();
      },
    );
  }

}

export default createPaginationContainer(
  JobListPage,
  graphql`
    fragment JobListPage_system on System
      @argumentDefinitions(
        count: {type: "Int", defaultValue: 20},
        cursor: {type: "String"},
        status: { type: "[JobStatus!]", defaultValue: null },
      ) {
      lastMessageId
      jobs(
       first: $count,
       after: $cursor,
       status: $status,
      )
        @connection(
          key: "JobListPage_jobs",
          filters: ["status"],
        ) {
        edges {
          node {
            ...JobTable_items
            id
          }
        }
      }
    }`,
  {
    direction: "forward",
    getConnectionFromProps: (props) => props.system && props.system.jobs,
    getVariables: (_, {count, cursor}, fragmentVariables) => ({
      count,
      cursor,
      status: fragmentVariables.status,
    }),
    query: graphql`
      query JobListPagePaginationQuery(
        $count: Int!,
        $cursor: String,
        $status: [JobStatus!],
      ) {
        system {
          ...JobListPage_system @arguments(
            count: $count,
            cursor: $cursor,
            status: $status,
          )
        }
      }
    `,
  },
);
