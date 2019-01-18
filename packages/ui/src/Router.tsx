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

function render({ Component, props }: RouteRenderArgs, error?: Error) {
  if (error) {
    return <ErrorPage error={error} />;
  }

  if (Component && props) {
    return <Component {...props} />;
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
          dataFrom="STORE_THEN_NETWORK"
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

  render: createRender({}),
});
