
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
  TextAreaProps,
} from "semantic-ui-react";

import "./SetKeyForm.css";

interface IProps {
  name: string;
  value: string;
  onChange: (obj: { name: string, value: string }) => any;
  onSubmit: () => any;
}

export default class SetKeyForm extends Component<IProps> {

  private nameRef: React.RefObject<HTMLInputElement>;
  private valueRef: React.RefObject<HTMLTextAreaElement>;

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
      <Form
        onSubmit={onSubmit}
        className="SetKeyForm"
      >
          <Form.Field>
            <label>Name</label>
            <Form.Input
              value={name}
              onChange={this.handleChangeName}
            >
              <input
                ref={this.nameRef}
              />
            </Form.Input>
          </Form.Field>
          <Form.Field>
            <label>Value</label>
            <textarea
              value={value}
              rows={8}
              ref={this.valueRef}
              onChange={this.handleChangeValue}
            />
          </Form.Field>
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

  private handleChangeName = (_: React.SyntheticEvent<HTMLElement>, { value }: InputProps) => {
    this.props.onChange({ ...this.props, name: value });
  }

  private handleChangeValue = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    this.props.onChange({ ...this.props, value: event.currentTarget.value });
  }

}
