
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
import {
  CheckboxProps,
  Form,
  TextAreaProps,
} from "semantic-ui-react";

import "./VariableFormField.css";

export interface IProps {
  name: string;
  value: string;
  save: boolean;
  onChangeValue: (value: string) => any;
  onChangeSave: (checked: boolean) => any;
}

export default class VariableFormField extends Component<IProps> {

  public render() {
    const { name, value, save } = this.props;

    return (
      <div className="VariableFormField">
        <Form.Field
        >
          <label>{name}</label>
          <Form.TextArea
            value={value}
            rows={8}
            onChange={this.handleChangeValue}
          />
        </Form.Field>
        <Form.Checkbox
          label="Save"
          checked={save}
          onChange={this.handleChangeCheckbox}
        />
      </div>
    );
  }

  private handleChangeValue = (_: React.SyntheticEvent<HTMLElement>, { value }: TextAreaProps) => {
    this.props.onChangeValue(value as string);
  }

  private handleChangeCheckbox = (_: React.SyntheticEvent<HTMLElement>, { checked }: CheckboxProps) => {
    this.props.onChangeSave(checked!);
  }

}