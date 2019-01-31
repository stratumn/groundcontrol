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

import React, { Component } from "react";
import { Radio } from "semantic-ui-react";

import "./LogEntryStatusFilter.css";

interface IProps {
  filters: string[] | undefined;
  onChange: (status: string[]) => any;
}

const allFilters = ["DEBUG", "INFO", "WARNING", "ERROR"];

// Note: we consider undefined filter to be the same as all status.
export default class LogEntryStatusFilter extends Component<IProps> {

  public render() {
    const filters = this.props.filters;
    const radios = allFilters.map((filter, i) => (
      <Radio
        key={i}
        label={filter}
        checked={!filters || filters.indexOf(filter) >= 0}
        onClick={this.handleToggleStatusFilter.bind(this, filter)}
      />
    ));

    return <div className="LogEntryStatusFilter">{radios}</div>;
  }

  private handleToggleStatusFilter(filter: string) {
    const filters = this.props.filters ?
      this.props.filters.slice() : allFilters.slice();
    const index = filters.indexOf(filter);

    if (index >= 0) {
      filters.splice(index, 1);
    } else {
      filters.push(filter);
    }

    this.props.onChange(filters);
  }

}
