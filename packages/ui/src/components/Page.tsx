
import React, { Component } from "react";
import {
  Header,
  Icon,
  SemanticICONS,
} from "semantic-ui-react";

import "./Page.css";

interface IProps {
  className?: string;
  icon: SemanticICONS;
  header: string;
  subheader: string;
}

export default class Page extends Component<IProps> {

  public render() {
    const { children, className, icon, header, subheader } = this.props;

    return (
      <div className={`Page ${className || ""}`}>
        <Header as="h1">
          <Icon name={icon} />
          <Header.Content>
            {header}
            <Header.Subheader>{subheader}</Header.Subheader>
          </Header.Content>
        </Header>
        {children}
      </div>
    );
  }

}
