
import debounce from "debounce";
import React, { Component } from "react";
import {
  Input,
  InputOnChangeData,
} from "semantic-ui-react";

interface IProps {
  onChange: (id: string) => any;
}

export default class WorkspaceSearch extends Component<IProps> {

  private handleSearchChange = debounce((_: React.ChangeEvent<HTMLInputElement>, { value }: InputOnChangeData) => {
    this.props.onChange(value);
  }, 100);

  public render() {
    return (
      <Input
        size="huge"
        icon="search"
        placeholder="Search..."
        style={{marginBottom: "2em"}}
        onChange={this.handleSearchChange}
      />
    );
  }

}
