import { Link, Router } from "found";
import React, { Component } from "react";
import {
  Header,
  Icon,
  Placeholder,
  Radio,
  Table,
 } from "semantic-ui-react";

import { RouterJobsListQueryResponse } from "../__generated__/RouterJobsListQuery.graphql";

import Moment from "react-moment";
import RepoShortName from "./RepoShortName";

interface IProps extends RouterJobsListQueryResponse {
  params: {
    status: string | undefined;
  };
  router: Router;
}

export default class JobsList extends Component<IProps> {

  public render() {
    const statuses = this.props.params.status ? this.props.params.status.split(",") : null;

    let rows: JSX.Element[];

    if (!this.props) {
      rows = [...Array(10)].map((_, i) => (
        <Table.Row key={i}>
          <Table.Cell><Placeholder><Placeholder.Line /></Placeholder></Table.Cell>
          <Table.Cell><Placeholder><Placeholder.Line /></Placeholder></Table.Cell>
          <Table.Cell><Placeholder><Placeholder.Line /></Placeholder></Table.Cell>
          <Table.Cell><Placeholder><Placeholder.Line /></Placeholder></Table.Cell>
          <Table.Cell><Placeholder><Placeholder.Line /></Placeholder></Table.Cell>
          <Table.Cell><Placeholder><Placeholder.Line /></Placeholder></Table.Cell>
        </Table.Row>
      ));
    } else {
      const jobs = this.props.jobs!.edges!;

      rows = jobs.map((edge) => {
        const node = edge!.node;

        return (
          <Table.Row key={node.id}>
            <Table.Cell>{node.name}</Table.Cell>
            <Table.Cell><Moment>{node.createdAt}</Moment></Table.Cell>
            <Table.Cell><Moment>{node.updatedAt}</Moment></Table.Cell>
            <Table.Cell>
              <Link to={`/workspaces/${node.project.workspace.slug}`}>{node.project.workspace.name}</Link>
            </Table.Cell>
            <Table.Cell><RepoShortName repo={node.project.repo} /></Table.Cell>
            <Table.Cell
              positive={node.status === "DONE"}
              warning={node.status === "RUNNING"}
              negative={node.status === "FAILED"}
            >
              {node.status}
            </Table.Cell>
          </Table.Row>
        );
      });
    }

    const radios = ["QUEUED", "RUNNING", "DONE", "FAILED"].map((status, i) => (
      <Radio
        key={i}
        label={status}
        checked={!statuses || statuses!.indexOf(status) >= 0}
        style={{marginRight: "2em"}}
        onClick={this.handleToggleStatus.bind(this, status)}
      />
    ));

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
        {radios}
        <Table celled={true}>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell>Name</Table.HeaderCell>
              <Table.HeaderCell>Created At</Table.HeaderCell>
              <Table.HeaderCell>Updated At</Table.HeaderCell>
              <Table.HeaderCell>Workspace</Table.HeaderCell>
              <Table.HeaderCell>Repo</Table.HeaderCell>
              <Table.HeaderCell>Status</Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {rows}
          </Table.Body>
        </Table>
      </div>
    );
  }

  private handleToggleStatus(status: string) {
    const statuses = this.props.params.status ?
      this.props.params.status.split(",") :
      ["QUEUED", "RUNNING", "DONE", "FAILED"];

    const index = statuses.indexOf(status);

    if (index >= 0) {
      statuses.splice(index, 1);
    } else {
      statuses.push(status);
    }

    if (statuses.length < 1) {
      return this.props.router.replace("/jobs");
    }

    this.props.router.replace(`/jobs/filter/${statuses.join(",")}`);
  }

}
