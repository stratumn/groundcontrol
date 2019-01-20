
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

import debounce from "debounce";
import React, { Component } from "react";
import {
  Input,
  InputOnChangeData,
} from "semantic-ui-react";

import "./WorkspaceSearch.css";

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
        className="WorkspaceSearch"
        size="huge"
        icon="search"
        placeholder="Search..."
        onChange={this.handleSearchChange}
      />
    );
  }

}
