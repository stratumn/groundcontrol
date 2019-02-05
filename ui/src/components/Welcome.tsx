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

import { Link } from "found";
import React, { Component } from "react";
import { Button, Message, Segment } from "semantic-ui-react";

import Page from "../components/Page";

export default class Welcome extends Component {

  public render() {
    return (
      <Page
        header="Welcome to Ground Control!"
        subheader="Hassle-free multi repository management."
        icon="rocket"
      >
        <Segment>
          <Message color="teal">
            <p>
              Before you can start working on cool apps, you need to add some sources.
              Sources contains workspaces which let you work on projects.
            </p>
            <p>
              <Link
                to="/sources"
                Component={Button}
                color="teal"
              >
                Let's add a source to get started!
              </Link>
            </p>
          </Message>
          <p>
            Ground Control is an application to help deal with multi-repository development
            using a user friendly web interface.
          </p>
          <p>
            Workspaces are defined using YAML files which can easily be shared.
            A workspace contains multiple projects.
            A project corresponds to a branch of a repository.
          </p>
          <p>
            The Ground Control user interface allows you to perform operations across the projects of a workspace,
            including:
          </p>
          <ul>
            <li>Clone all repositories (defaults to $HOME/groundcontrol/workspaces/WORKSPACE/PROJECT)</li>
            <li>Check the status of repositories against their origins</li>
            <li>Pull all outdated repositories</li>
            <li>Define workspace wide tasks</li>
            <li>Create scripts to launch multi-repository applications</li>
          </ul>
          <p>
            Good luck and happy coding!
          </p>
        </Segment>
      </Page>
    );
  }
}
