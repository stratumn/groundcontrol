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

import App from "./components/App";
import WorkspacesList from "./components/WorkspacesList";
import WorkspacesView from "./components/WorkspacesView";
import JobsPage from "./containers/JobsPage";

const workspacesListQuery = graphql`
  query RouterWorkspacesListQuery {
    viewer {
      workspaces {
        id
        name
        slug
        description
        projects {
          id
          repo
          branch
        }
      }
    }
  }
`;

const workspacesViewQuery = graphql`
  query RouterWorkspacesViewQuery($slug: String!) {
    viewer {
      workspace(slug: $slug) {
        name
        slug
        description
        notes
        projects {
          id
          repo
          branch
          description
          commits(first: 3) {
            edges {
              node {
                id
                headline
                date
                author
              }
            }
          }
        }
      }
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
          Component={WorkspacesList}
          query={workspacesListQuery}
        />
        <Route
          path=":slug"
          Component={WorkspacesView}
          query={workspacesViewQuery}
        />
      </Route>
      <Route path="jobs">
        <Route
          Component={JobsPage}
          query={jobsQuery}
        />
        <Route
          path="filter/:filters"
          Component={JobsPage}
          query={jobsQuery}
          prepareVariables={prepareJobsVariables}
        />
      </Route>
    </Route>,
  ),

  render: createRender({}),
});
