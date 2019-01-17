
import React, { Component } from "react";
import {
  Header,
  Icon,
  SemanticICONS,
} from "semantic-ui-react";

interface IProps {
  icon: SemanticICONS;
  header: string;
  subheader: string;
}

export default class Page extends Component<IProps> {

  public render() {
    const { children, icon, header, subheader } = this.props;

    return (
      <div>
        <Header as="h1" style={{marginBottom: "1em"}}>
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
