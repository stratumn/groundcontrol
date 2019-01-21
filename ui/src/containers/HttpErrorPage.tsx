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

import { HttpError } from "found";
import React, { Component } from "react";

import Page from "../components/Page";

interface IProps {
  error: HttpError;
}

export default class ErrorPage extends Component<IProps> {

  public render() {
    const error = this.props.error;

    return (
      <Page
        header="Oops"
        subheader="Looks like something's wrong."
        icon="warning"
      >
        <h4>Error {error.status}</h4>
        <pre>{error.data}</pre>
      </Page>
    );
  }

}