
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

import React, { Component} from "react";
import {
  Button,
  Form,
  InputProps,
} from "semantic-ui-react";

import "./Page.css";

interface IProps {
  onSet: (name: string, value: string) => any;
}

interface IState {
  name: string;
  value: string;
}

export default class SetKeyForm extends Component<IProps, IState> {

  public state: IState = {
    name: "",
    value: "",
  };

  public render() {
    const { name, value } = this.state;
    const disabled = !name;

    return (
      <Form onSubmit={this.handleSubmit}>
        <Form.Group>
          <Form.Field width="5">
            <label>Name</label>
            <Form.Input
              name="name"
              value={name}
              onChange={this.handleChangeInput}
            />
          </Form.Field>
          <Form.Field width="12">
            <label>Value</label>
            <Form.Input
              name="value"
              value={value}
              onChange={this.handleChangeInput}
            />
          </Form.Field>
        </Form.Group>
        <Button
          type="submit"
          color="teal"
          icon="edit"
          content="Set"
          disabled={disabled}
        />
      </Form>
    );
  }

  private handleChangeInput = (_: React.SyntheticEvent<HTMLElement>, { name, value }: InputProps) => {
    switch (name) {
    case "name": this.setState({ name: value }); break;
    case "value": this.setState({ value }); break;
    }
  }

  private handleSubmit = () => {
    this.props.onSet(this.state.name, this.state.value);
    this.setState({ name: "", value: "" });
  }

}
