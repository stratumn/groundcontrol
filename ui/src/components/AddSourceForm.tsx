
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

import React, { Component, Fragment } from "react";
import {
  Button,
  DropdownProps,
  Form,
  InputProps,
} from "semantic-ui-react";

import "./Page.css";

interface IProps {
  onAddDirectorySource: (directory: string) => any;
  onAddGitSource: (repository: string, branch: string) => any;
}

interface IState {
  type: string;
  directory: string;
  repository: string;
  branch: string;
}

export default class AddSourceForm extends Component<IProps, IState> {

  public state: IState = {
    branch: "",
    directory: "",
    repository: "",
    type: "directory",
  };

  public render() {
    const { type, directory, repository, branch } = this.state;
    const options = [
      { key: "directory", text: "Directory", value: "directory" },
      { key: "git", text: "Git", value: "git" },
    ];
    const disabled = type === "directory" && !directory || type === "git" && !repository;

    let typeFields: JSX.Element;

    if (type === "directory") {
      typeFields = (
        <Form.Field width="13">
          <label>Directory (absolute path)</label>
          <Form.Input
            name="directory"
            value={directory}
            onChange={this.handleChangeInput}
          />
        </Form.Field>
      );
    } else {
      typeFields = (
        <Fragment>
          <Form.Field width="9">
            <label>Repository</label>
            <Form.Input
              name="repository"
              value={repository}
              onChange={this.handleChangeInput}
            />
          </Form.Field>
          <Form.Field width="4">
            <label>Branch</label>
            <Form.Input
              name="branch"
              placeholder="master"
              value={branch}
              onChange={this.handleChangeInput}
            />
          </Form.Field>
        </Fragment>
      );
    }

    return (
      <Form onSubmit={this.handleSubmit}>
        <Form.Group>
          <Form.Select
            label="Type"
            options={options}
            defaultValue="directory"
            width="3"
            onChange={this.handleChangeType}
          />
          {typeFields}
        </Form.Group>
        <Button
          type="submit"
          color="teal"
          icon="add"
          content="Add"
          disabled={disabled}
        />
      </Form>
    );
  }

  private handleChangeType = (_: React.SyntheticEvent<HTMLElement>, { value }: DropdownProps) => {
    this.setState({ type: value as string });
  }

  private handleChangeInput = (_: React.SyntheticEvent<HTMLElement>, { name, value }: InputProps) => {
    switch (name) {
    case "directory": this.setState({ directory: value }); break;
    case "repository": this.setState({ repository: value }); break;
    case "branch": this.setState({ branch: value }); break;
    }
  }

  private handleSubmit = () => {
    if (this.state.type === "directory") {
      this.props.onAddDirectorySource(this.state.directory);
      this.setState({ directory: "" });
    } else {
      this.props.onAddGitSource(this.state.repository, this.state.branch || "master");
      this.setState({ repository: "", branch: "" });
    }
  }

}
