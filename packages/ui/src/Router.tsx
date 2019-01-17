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
        />
        <Route
          path=":slug"
          Component={WorkspaceViewPage}
          query={workspaceViewQuery}
        />
      </Route>
      <Route path="jobs">
        <Route
          Component={JobListPage}
          query={jobListQuery}
        />
        <Route
          path=":filters"
          Component={JobListPage}
          query={jobListQuery}
          prepareVariables={prepareJobListVariables}
        />
      </Route>
    </Route>,
  ),

  render: createRender({}),
});
