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
import { Disposable } from "relay-runtime";
import { Button } from "semantic-ui-react";

import { createPaginationContainer, RelayPaginationProp } from "react-relay";

import { JobListPage_viewer } from "./__generated__/JobListPage_viewer.graphql";

import JobFilter from "../components/JobFilter";
import JobTable from "../components/JobTable";
import Page from "../components/Page";

import { subscribe } from "../subscriptions/jobUpserted";

interface IProps {
  relay: RelayPaginationProp;
  router: Router;
  viewer: JobListPage_viewer;
  params: {
    filters: string | undefined;
  };
}

export class JobListPage extends Component<IProps> {

  private disposables: Disposable[] = [];

  public render() {
    const items = this.props.viewer.jobs!.edges.map(({ node }) => node);
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
        <JobTable items={items} />
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
    this.disposables.push(subscribe(this.props.relay.environment));
  }

  public componentWillUnmount() {
    for (const disposable of this.disposables) {
      disposable.dispose();
    }

    this.disposables = [];
  }

  private handleFiltersChange = (filters: string[]) => {
    if (filters.length < 1 || filters.length > 3) {
      return this.props.router.replace("/jobs");
    }

    this.props.router.replace(`/jobs/${filters.join(",")}`);
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

}

export default createPaginationContainer(
  JobListPage,
  graphql`
    fragment JobListPage_viewer on User
      @argumentDefinitions(
        count: {type: "Int", defaultValue: 10},
        cursor: {type: "String"},
        status: { type: "[JobStatus!]", defaultValue: null },
      ) {
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
    getConnectionFromProps: (props) => props.viewer && props.viewer.jobs,
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
        viewer {
          ...JobListPage_viewer @arguments(
            count: $count,
            cursor: $cursor,
            status: $status,
          )
        }
      }
    `,
  },
);
