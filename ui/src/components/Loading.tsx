
import React, { Component } from "react";
import { Loader } from "semantic-ui-react";

import "./Loading.css";

export default class Loading extends Component {

  public render() {
    return (
      <div className="Loading">
        <Loader active={true} size="massive" />
      </div>
    );
  }

}
