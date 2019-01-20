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
import { BrowserProtocol, queryMiddleware } from "farce";
import {
  createFarceRouter,
  createRender,
  makeRouteConfig,
  Redirect,
  Route,
  RouteRenderArgs,
} from "found";
import React from "react";

import Loading from "./components/Loading";
import App from "./containers/App";
import ErrorPage from "./containers/ErrorPage";
import HttpErrorPage from "./containers/HttpErrorPage";
import JobListPage from "./containers/JobListPage";
import WorkspaceListPage from "./containers/WorkspaceListPage";
import WorkspaceViewPage from "./containers/WorkspaceViewPage";

const workspaceListQuery = graphql`
  query RouterWorkspacesQuery {
    viewer {
      ...WorkspaceListPage_viewer
    }
  }
`;

const workspaceViewQuery = graphql`
  query RouterWorkspaceQuery($slug: String!) {
    viewer {
      ...WorkspaceViewPage_viewer @arguments(slug: $slug)
    }
  }
`;

const jobListQuery = graphql`
  query RouterJobsQuery($status: [JobStatus!]) {
    viewer {
      ...JobListPage_viewer @arguments(status: $status)
    }
  }
`;

function prepareJobListVariables({ filters }: { filters: string }) {
  return { status: filters ? filters.split(",") : null };
}

function render(args: RouteRenderArgs) {
  // Only way I could find to get relay errors :(
  const error = ((args as any).error);
  if (error) {
    return <ErrorPage error={error} />;
  }

  if (args.Component && args.props) {
    return <args.Component {...args.props} />;
  }

  return <Loading />;
}

export default createFarceRouter({
  historyMiddlewares: [queryMiddleware],
  historyProtocol: new BrowserProtocol(),
  routeConfig: makeRouteConfig(
    <Route
      path="/"
      Component={App}
    >
      <Redirect from="/" to="/workspaces" />
      <Route path="workspaces">
        <Route
          Component={WorkspaceListPage}
          query={workspaceListQuery}
          render={render}
        />
        <Route
          path=":slug"
          Component={WorkspaceViewPage}
          query={workspaceViewQuery}
          render={render}
        />
      </Route>
      <Route path="jobs">
        <Route
          Component={JobListPage}
          query={jobListQuery}
          render={render}
        />
        <Route
          path=":filters"
          Component={JobListPage}
          query={jobListQuery}
          prepareVariables={prepareJobListVariables}
          render={render}
        />
      </Route>
    </Route>,
  ),

  render: createRender({
    renderError: ({ error }) => (
      <App>
        <HttpErrorPage error={error} />
      </App>
    ),
  }),

});
