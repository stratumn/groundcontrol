import { HttpError } from "found";
import React, { Component } from "react";

import Page from "../components/Page";

interface IProps {
  error: Error;
}

export default class ErrorPage extends Component<IProps> {

  public render() {
    const error = this.props.error;

    return (
      <Page
        header="Oops"
        subheader="Looks like something's wrong."
        icon="warning"
      >
        <pre>{error.stack}</pre>
      </Page>
    );
  }

}