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
import JobsList from "./components/JobsList";
import WorkspacesList from "./components/WorkspacesList";
import WorkspacesView from "./components/WorkspacesView";

const workspacesListQuery = graphql`
  query RouterWorkspacesListQuery {
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
`;

const workspacesViewQuery = graphql`
  query RouterWorkspacesViewQuery($slug: String!) {
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
`;

const jobsListQuery = graphql`
  query RouterJobsListQuery($status: [JobStatus!]) {
    jobs(status: $status) {
      edges {
        node {
          id
          name
          status
          createdAt
          updatedAt
          project {
            repo
            branch
            workspace {
              slug
              name
            }
          }
        }
      }
    }
  }
`;

function prepareJobsListStatusVariables({ status }: { status: string }) {
  return { status: status ? status.split(",") : [] };
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
          Component={JobsList}
          query={jobsListQuery}
        />
        <Route
          path="filter/:status?"
          Component={JobsList}
          query={jobsListQuery}
          prepareVariables={prepareJobsListStatusVariables}
        />
      </Route>
    </Route>,
  ),

  render: createRender({}),
});
