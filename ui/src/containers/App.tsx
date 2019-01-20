import React, { Component } from "react";
import { Container } from "semantic-ui-react";

import Menu from "../components/Menu";

import "./App.css";

export default class App extends Component {
  public render() {
    return (
      <div className="App">
        <Menu />
        <Container>
          {this.props.children}
        </Container>
      </div>
    );
  }
}
