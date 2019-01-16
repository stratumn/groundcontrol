import React, { Component } from "react";

interface IProps {
  repo: string;
}

export default class RepoShortName extends Component<IProps> {

  public render() {
    const shortName = this.props.repo
      .replace(/^git@github\.com:/, "")
      .replace(/\.git$/, "");

    return <span>{shortName}</span>;
  }
}
