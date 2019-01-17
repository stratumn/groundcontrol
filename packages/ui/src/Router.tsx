import graphql from "babel-plugin-relay/macro";
import { BrowserProtocol, queryMiddleware } from "farce";
import {
  createFarceRouter,
  createRender,
  makeRouteConfig,
  Redirect,
  Route,
} from "found";
import React from "react";

import App from "./containers/App";
import JobsPage from "./containers/JobsPage";
import WorkspacePage from "./containers/WorkspacePage";
import WorkspacesPage from "./containers/WorkspacesPage";

const workspacesQuery = graphql`
  query RouterWorkspacesQuery {
    viewer {
      ...WorkspacesPage_viewer
    }
  }
`;

const workspaceQuery = graphql`
  query RouterWorkspaceQuery($slug: String!) {
    viewer {
      ...WorkspacePage_viewer @arguments(slug: $slug)
    }
  }
`;

const jobsQuery = graphql`
  query RouterJobsQuery($status: [JobStatus!]) {
    viewer {
      ...JobsPage_viewer @arguments(status: $status)
    }
  }
`;

function prepareJobsVariables({ filters }: { filters: string }) {
  return { status: filters ? filters.split(",") : null };
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
          Component={WorkspacesPage}
          query={workspacesQuery}
        />
        <Route
          path=":slug"
          Component={WorkspacePage}
          query={workspaceQuery}
        />
      </Route>
      <Route path="jobs">
        <Route
          Component={JobsPage}
          query={jobsQuery}
        />
        <Route
          path="filter/:filters"
          Component={WorkspacePage}
          query={workspaceQuery}
        />
      </Route>
    </Route>,
  ),

  render: createRender({}),
});
