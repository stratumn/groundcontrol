
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
  name: string;
  value: string;
  onChange: (obj: { name: string, value: string }) => any;
  onSubmit: () => any;
}

export default class SetKeyForm extends Component<IProps> {

  private nameRef: React.RefObject<HTMLInputElement>;
  private valueRef: React.RefObject<HTMLInputElement>;

  private shouldFocusName = false;
  private shouldFocusValue = false;

  constructor(props: IProps) {
    super(props);
    this.nameRef = React.createRef();
    this.valueRef = React.createRef();
  }

  public render() {
    const { name, value, onSubmit } = this.props;
    const disabled = !name;

    return (
      <Form onSubmit={onSubmit}>
        <Form.Group>
          <Form.Field width="5">
            <label>Name</label>
            <Form.Input
              name="name"
              value={name}
              onChange={this.handleChangeInput}
            >
              <input
                ref={this.nameRef}
              />
            </Form.Input>
          </Form.Field>
          <Form.Field width="11">
            <label>Value</label>
            <Form.Input
                name="value"
                value={value}
                onChange={this.handleChangeInput}
            >
              <input
                ref={this.valueRef}
              />
            </Form.Input>
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

  public componentDidUpdate() {
    if (this.shouldFocusName) {
      const input = this.nameRef.current;

      if (input) {
        input.focus();
        input.select();
      }
    }

    if (this.shouldFocusValue) {
      const input = this.valueRef.current;

      if (input) {
        input.focus();
        input.select();
      }
    }

    this.shouldFocusName = false;
    this.shouldFocusValue = false;
  }

  public selectName() {
    this.shouldFocusName = true;
  }

  public selectValue() {
    this.shouldFocusValue = true;
  }

  private handleChangeInput = (_: React.SyntheticEvent<HTMLElement>, { name, value }: InputProps) => {
    const onChange = this.props.onChange;
    const obj = { ...this.props };

    switch (name) {
    case "name":
    case "value":
      onChange({ ...obj, [name]: value });
      break;
    }
  }

}
