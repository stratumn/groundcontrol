import React, { Component } from "react";

interface IProps {
  repository: string;
}

export default class RepositoryShortName extends Component<IProps> {

  public render() {
    const shortName = this.props.repository
      .replace(/^git@github\.com:/, "")
      .replace(/\.git$/, "");

    return <span>{shortName}</span>;
  }
}
