import React, { Component } from "react";
import { Radio } from "semantic-ui-react";

interface IProps {
  filters: string[];
  onChange: (status: string[]) => any;
}

const allFilters = ["QUEUED", "RUNNING", "DONE", "FAILED"];

// Note: we consider no status to be the same as all status.
export default class JobsListFilter extends Component<IProps> {

  public render() {
    const filters = this.props.filters;

    return allFilters.map((filter, i) => (
      <Radio
        key={i}
        label={filter}
        checked={filters.length < 1 || filters.indexOf(filter) >= 0}
        style={{marginRight: "2em"}}
        onClick={this.handleToggleFilter.bind(this, filter)}
      />
    ));
  }

  private handleToggleFilter(filter: string) {
    const filters = this.props.filters.length ?
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
