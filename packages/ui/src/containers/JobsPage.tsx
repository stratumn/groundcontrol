import graphql from "babel-plugin-relay/macro";
import { Router } from "found";
import React, { Component } from "react";
import { Disposable } from "relay-runtime";
import {
  Header,
  Icon,
 } from "semantic-ui-react";

import { createFragmentContainer } from "react-relay";

import { JobsPage_viewer } from "./__generated__/JobsPage_viewer.graphql";

import JobsList from "../components/JobsList";
import JobsListFilter from "../components/JobsListFilter";

import { subscribe } from "../subscriptions/jobUpserted";

interface IProps {
  viewer: JobsPage_viewer;
  params: {
    filters: string | undefined;
  };
  router: Router;
}

export class JobsPage extends Component<IProps> {

  private disposables: Disposable[] = [];

  public render() {
    const items = this.props.viewer.jobs!.edges.map((edge) => edge.node);
    const filters = this.props.params.filters === undefined ? undefined :
      this.props.params.filters.split(",");

    return (
      <div>
        <Header as="h1" style={{marginBottom: "2em"}}>
          <Icon name="tasks" />
          <Header.Content>
            Jobs
            <Header.Subheader>
              Jobs are short lived tasks such as cloning a repository.
            </Header.Subheader>
          </Header.Content>
        </Header>
        <JobsListFilter
          filters={filters}
          onChange={this.handleFiltersChange}
        />
        <JobsList items={items} />
      </div>
    );
  }

  public componentDidMount() {
    this.disposables.push(subscribe());
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

    this.props.router.replace(`/jobs/filter/${filters.join(",")}`);
  }

}

export default createFragmentContainer(JobsPage, graphql`
  fragment JobsPage_viewer on User
    @argumentDefinitions(
      status: { type: "[JobStatus!]", defaultValue: null },
    ) {
    jobs(first: 100, status: $status)
      @connection(
        key: "JobsPage_jobs",
        filters: ["status"],
      ) {
      edges {
        node {
          ...JobsList_items
          id
        }
      }
    }
  }`,
);
