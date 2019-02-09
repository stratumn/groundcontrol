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
import React, { Fragment } from "react";
import { createFragmentContainer } from "react-relay";
import {
  Container,
  Menu as SemanticMenu,
  Responsive,
  Sidebar,
} from "semantic-ui-react";

import { Menu_system } from "./__generated__/Menu_system.graphql";

import MenuPrimaryItems from "./MenuPrimaryItems";
import MenuSecondaryItems from "./MenuSecondaryItems";

interface IProps {
  system: Menu_system;
  showSidebar: boolean;
  onShowSidebar: () => any;
  onHideSidebar: () => any;
}

export const Menu = ({
  system: {
    jobMetrics,
    processMetrics,
    logMetrics,
  },
  showSidebar,
  onShowSidebar,
  onHideSidebar,
}: IProps) => (
  <Fragment>
    <SemanticMenu
      fixed="top"
      size="large"
      color="teal"
      inverted={true}
    >
      <Container fluid={true}>
        <Responsive
          as={SemanticMenu.Item}
          icon="bars"
          maxWidth={767}
          onClick={onShowSidebar}
        />
        <Responsive
          as={MenuPrimaryItems}
          minWidth={768}
        />
        <Responsive
          as={SemanticMenu.Menu}
          position="right"
          minWidth={768}
        >
          <MenuSecondaryItems
            jobMetrics={jobMetrics}
            processMetrics={processMetrics}
            logMetrics={logMetrics}
          />
        </Responsive>
      </Container>
    </SemanticMenu>
    <Responsive maxWidth={767}>
      <Sidebar
        as={SemanticMenu}
        vertical={true}
        animation="push"
        visible={showSidebar}
        onHide={onHideSidebar}
      >
        <MenuPrimaryItems />
        <MenuSecondaryItems
          jobMetrics={jobMetrics}
          processMetrics={processMetrics}
          logMetrics={logMetrics}
        />
      </Sidebar>
    </Responsive>
  </Fragment>
);

export default createFragmentContainer(Menu, graphql`
  fragment Menu_system on System {
    jobMetrics {
      ...MenuSecondaryItems_jobMetrics
    }
    processMetrics {
      ...MenuSecondaryItems_processMetrics
    }
    logMetrics {
      ...MenuSecondaryItems_logMetrics
    }
  }`,
);
