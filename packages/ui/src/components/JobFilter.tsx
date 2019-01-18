import React, { Component } from "react";
import { Radio } from "semantic-ui-react";

import "./JobFilter.css";

interface IProps {
  filters: string[] | undefined;
  onChange: (status: string[]) => any;
}

const allFilters = ["QUEUED", "RUNNING", "DONE", "FAILED"];

// Note: we consider undefined filter to be the same as all status.
export default class JobFilter extends Component<IProps> {

  public render() {
    const filters = this.props.filters;
    const radios = allFilters.map((filter, i) => (
      <Radio
        key={i}
        label={filter}
        checked={!filters || filters.indexOf(filter) >= 0}
        onClick={this.handleToggleFilter.bind(this, filter)}
      />
    ));

    return <div className="JobFilter">{radios}</div>;
  }

  private handleToggleFilter(filter: string) {
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
