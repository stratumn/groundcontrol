import graphql from "babel-plugin-relay/macro";
import { Router } from "found";
import React, { Component } from "react";
import { Disposable } from "relay-runtime";
import {
  Header,
  Icon,
 } from "semantic-ui-react";

import { createFragmentContainer } from "react-relay";

import { JobListPage_viewer } from "./__generated__/JobListPage_viewer.graphql";

import JobFilter from "../components/JobFilter";
import JobList from "../components/JobList";

import { subscribe } from "../subscriptions/jobUpserted";

interface IProps {
  viewer: JobListPage_viewer;
  params: {
    filters: string | undefined;
  };
  router: Router;
}

export class JobListPage extends Component<IProps> {

  private disposables: Disposable[] = [];

  public render() {
    const items = this.props.viewer.jobs!.edges.map(({ node }) => node);
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
        <JobFilter
          filters={filters}
          onChange={this.handleFiltersChange}
        />
        <JobList items={items} />
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

    this.props.router.replace(`/jobs/${filters.join(",")}`);
  }

}

export default createFragmentContainer(JobListPage, graphql`
  fragment JobListPage_viewer on User
    @argumentDefinitions(
      status: { type: "[JobStatus!]", defaultValue: null },
    ) {
    jobs(first: 100, status: $status)
      @connection(
        key: "JobListPage_jobs",
        filters: ["status"],
      ) {
      edges {
        node {
          ...JobList_items
          id
        }
      }
    }
  }`,
);
