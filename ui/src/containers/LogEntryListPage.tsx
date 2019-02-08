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
import { Container, Loader } from "semantic-ui-react";

import { LogEntryListPage_system } from "./__generated__/LogEntryListPage_system.graphql";
import { LogEntryListPage_viewer } from "./__generated__/LogEntryListPage_viewer.graphql";

import LogEntryFilter from "../components/LogEntryFilter";
import LogEntryTable from "../components/LogEntryTable";

import { subscribe } from "../subscriptions/logEntryAdded";

import "./LogEntryListPage.css";

interface IProps {
  relay: RelayPaginationProp;
  router: Router;
  viewer: LogEntryListPage_viewer;
  system: LogEntryListPage_system;
  params: {
    filters: string | undefined;
    ownerId: string | undefined;
  };
}

export class LogEntryListPage extends Component<IProps> {

  private disposables: Disposable[] = [];

  private prevScrollY = 0;
  private previousScrollHeight = 0;
  private shouldScrollDown = false;

  public render() {
    const items = this.props.system.logEntries.edges.map(({ node }) => node);
    const filters = this.props.params.filters === undefined ? undefined :
      this.props.params.filters.split(",");
    const ownerId = this.props.params.ownerId;
    const projects = this.props.viewer.projects.edges.map(({ node }) => node);
    const isLoading = this.props.relay.isLoading();

    return (
      <Container
        className="LogEntryListPage"
        fluid={true}
      >
        <Loader
          inverted={true}
          active={isLoading}
        />
        <LogEntryFilter
          filters={filters}
          projects={projects}
          ownerId={ownerId}
          onChange={this.handleFiltersChange}
        />
        <LogEntryTable items={items} />
      </Container>
    );
  }

  public componentDidMount() {
    const environment = this.props.relay.environment;
    const lastMessageId = this.props.system.lastMessageId;
    this.disposables.push(subscribe(environment, lastMessageId));

    window.scrollTo(0, document.body.scrollHeight);

    document.addEventListener("scroll", this.handleScroll);
    this.disposables.push({
      dispose: () => {
        document.removeEventListener("scroll", this.handleScroll);
      },
    });

    this.previousScrollHeight = document.body.scrollHeight;
  }

  public componentDidUpdate() {
    if (this.shouldScrollDown) {
      window.scrollTo(0, document.body.scrollHeight);
    }

    this.previousScrollHeight = document.body.scrollHeight;
  }

  public componentWillUnmount() {
    for (const disposable of this.disposables) {
      disposable.dispose();
    }

    this.disposables = [];
  }

  private handleFiltersChange = (filters: string[], ownerId?: string) => {
    if ((filters.length < 1 || filters.length > 3) && !ownerId) {
      return this.props.router.replace("/logs");
    }

    this.props.router.replace(`/logs/${filters.join(",")};${ownerId || ""}`);
  }

  private handleScroll = () => {
    this.shouldScrollDown = window.scrollY + window.innerHeight === document.body.scrollHeight;

    const prevScrollY = this.prevScrollY;
    this.prevScrollY = window.scrollY;
    if (prevScrollY < window.scrollY) {
      return;
    }

    if (!this.props.relay.hasMore()) {
      return;
    }

    if (this.props.relay.isLoading()) {
      return;
    }

    if (window.scrollY > window.innerHeight) {
      return;
    }

    this.previousScrollHeight = document.body.scrollHeight;

    const disposable = this.props.relay.loadMore(
      100,
      (err) => {
        if (err) {
          console.log(err);
        }

        if (this.previousScrollHeight < document.body.scrollHeight) {
          window.scrollTo(0, window.scrollY + document.body.scrollHeight - this.previousScrollHeight);
        }

        this.previousScrollHeight = document.body.scrollHeight;

        // To hide loader.
        this.forceUpdate();
      },
    );

    if (disposable) {
      this.disposables.push(disposable);
    }

    // To show loader.
    this.forceUpdate();
  }

}

export default createPaginationContainer(
  LogEntryListPage,
  graphql`
    fragment LogEntryListPage_system on System
      @argumentDefinitions(
        count: {type: "Int", defaultValue: 100},
        cursor: {type: "String"},
        level: { type: "[LogLevel!]", defaultValue: null },
        ownerId: { type: "ID", defaultValue: null }
      ) {
      lastMessageId
      logEntries(
       last: $count,
       before: $cursor,
       level: $level,
       ownerId: $ownerId,
      )
        @connection(
          key: "LogEntryListPage_logEntries",
          filters: ["level", "ownerId"],
        ) {
        edges {
          node {
            ...LogEntryTable_items
            id
          }
        }
      }
    }
    fragment LogEntryListPage_viewer on User {
      projects {
        edges {
          node {
            ...LogEntryFilter_projects
          }
        }
      }
    }`,
  {
    direction: "backward",
    getConnectionFromProps: (props) => props.system && props.system.logEntries,
    getVariables: (_, {count, cursor}, fragmentVariables) => ({
      count,
      cursor,
      level: fragmentVariables.level,
      ownerId: fragmentVariables.ownerId,
    }),
    query: graphql`
      query LogEntryListPagePaginationQuery(
        $count: Int!,
        $cursor: String,
        $level: [LogLevel!],
        $ownerId: ID,
      ) {
        system {
          ...LogEntryListPage_system @arguments(
            count: $count,
            cursor: $cursor,
            level: $level,
            ownerId: $ownerId,
          )
        }
      }
    `,
  },
);
