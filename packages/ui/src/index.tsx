import { Resolver } from "found-relay";
import React from "react";
import ReactDOM from "react-dom";

import groundcontrol from "./groundcontrol.env.relay";
import Router from "./Router";
import * as serviceWorker from "./serviceWorker";

import jobUpserted from "./subscriptions/jobUpserted";

import "./index.css";

ReactDOM.render(
  <Router resolver={new Resolver(groundcontrol)} />,
  document.getElementById("root"),
);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: http://bit.ly/CRA-PWA
serviceWorker.unregister();

jobUpserted();
