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
import React, { Component } from "react";
import { createFragmentContainer } from "react-relay";

import { LogEntryFilter_projects } from "./__generated__/LogEntryFilter_projects.graphql";

import LogEntryOwnerFilter from "./LogEntryOwnerFilter";
import LogEntryStatusFilter from "./LogEntryStatusFilter";

interface IProps {
  filters?: string[];
  projects: LogEntryFilter_projects;
  ownerId?: string;
  onChange: (status: string[], ownerID?: string) => any;
}

export class LogEntryFilter extends Component<IProps> {

  public render() {
    const { projects, ownerId, filters } = this.props;

    return (
      <div>
        <LogEntryOwnerFilter
          items={projects}
          ownerId={ownerId}
          onChange={this.handleOwnerChange}
        />
        <LogEntryStatusFilter
          filters={filters}
          onChange={this.handleStatusChange}
        />
      </div>
    );
  }

  private handleOwnerChange = (ownerId?: string) => {
    this.props.onChange(this.props.filters || [], ownerId);
  }

  private handleStatusChange = (status: string[]) => {
    this.props.onChange(status, this.props.ownerId);
  }

}

export default createFragmentContainer(LogEntryFilter, graphql`
  fragment LogEntryFilter_projects on Project @relay(plural: true) {
    ...LogEntryOwnerFilter_items
  }`,
);
