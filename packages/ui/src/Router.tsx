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

const workspacesListQuery = graphql`
  query RouterWorkspacesListQuery {
    workspaces {
      name
      slug
      description
      projects {
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
    </Route>,
  ),

  render: createRender({}),
});
