
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
  Button,
  Modal,
} from "semantic-ui-react";

import VariableForm, { IProps as IFormProps } from "./VariableForm";

interface IProps extends IFormProps {
  onClose: () => any;
}

export default class VariableFormModal extends Component<IProps> {

  public render() {
    const { onClose, onSubmit } = this.props;

    return (
      <Modal
        open={true}
        onClose={onClose}
      >
        <Modal.Header>Task Variables</Modal.Header>
        <Modal.Content scrolling={true}>
          <VariableForm {...this.props} />
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={onClose}>
            Cancel
          </Button>
          <Button
            icon="rocket"
            content="Proceed"
            color="teal"
            onClick={onSubmit}
          />
        </Modal.Actions>
      </Modal>
    );
  }

}
