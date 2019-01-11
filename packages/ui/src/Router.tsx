import { BrowserProtocol, queryMiddleware } from "farce";
import {
    createFarceRouter,
    createRender,
    makeRouteConfig,
    Redirect,
    Route,
  } from "found";
import graphql from "babel-plugin-relay/macro";
import React from "react";

import App from "./components/App";
import WorkspacesList from "./components/WorkspacesList";
import WorkspacesView from "./components/WorkspacesView";

export default createFarceRouter({
  historyProtocol: new BrowserProtocol(),
  historyMiddlewares: [queryMiddleware],
  routeConfig: makeRouteConfig(
    <Route
      path="/"
      Component={App}
    >
      <Redirect from="/" to="/workspaces" />
      <Route path="workspaces">
        <Route
          Component={WorkspacesList}
          query={graphql`
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
          `}
        />
        <Route
          path=":slug"
          Component={WorkspacesView}
        />
      </Route>
    </Route>,
  ),

  render: createRender({}),
});