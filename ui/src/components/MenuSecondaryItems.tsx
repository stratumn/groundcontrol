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

import graphql from "babel-plugin-relay/macro";
import { Link } from "found";
import React, { Fragment } from "react";
import { createFragmentContainer } from "react-relay";
import {
  Label,
  Menu,
} from "semantic-ui-react";

import { MenuSecondaryItems_jobMetrics } from "./__generated__/MenuSecondaryItems_jobMetrics.graphql";
import { MenuSecondaryItems_logMetrics } from "./__generated__/MenuSecondaryItems_logMetrics.graphql";
import { MenuSecondaryItems_processMetrics } from "./__generated__/MenuSecondaryItems_processMetrics.graphql";

interface IProps {
  jobMetrics: MenuSecondaryItems_jobMetrics;
  processMetrics: MenuSecondaryItems_processMetrics;
  logMetrics: MenuSecondaryItems_logMetrics;
}

export const MenuSecondaryItems = ({ jobMetrics, processMetrics, logMetrics }: IProps) => (
  <Fragment>
    <Link
      to="/jobs"
      Component={Menu.Item}
      activePropName="active"
    >
      Jobs
      <Label color="blue">
        {jobMetrics.queued + jobMetrics.running}
      </Label>
    </Link>
    <Link
      to="/processes"
      Component={Menu.Item}
      activePropName="active"
    >
      Processes
      <Label color="blue">
        {processMetrics.running}
      </Label>
    </Link>
    <Link
      to="/logs"
      Component={Menu.Item}
      activePropName="active"
    >
      Logs
      <Label color="pink">
        {logMetrics.error}
      </Label>
    </Link>
    <Menu.Item href="http://localhost:3333/graphql">
      GraphQL
    </Menu.Item>
  </Fragment>
);

export default createFragmentContainer(MenuSecondaryItems, graphql`
  fragment MenuSecondaryItems_jobMetrics on JobMetrics {
    queued
    running
  }
  fragment MenuSecondaryItems_processMetrics on ProcessMetrics {
    running
  }
  fragment MenuSecondaryItems_logMetrics on LogMetrics {
    error
  }
`);
