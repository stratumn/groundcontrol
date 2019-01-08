import React, { Component } from 'react';
import {QueryRenderer} from 'react-relay';
/// <reference path="declarations.d.ts"/>
import graphql from 'babel-plugin-relay/macro';
import {Github} from './relay';
import { Container, Card, Placeholder, List } from 'semantic-ui-react';
import './App.css';

class App extends Component {
  render() {
    return (
      <QueryRenderer
        environment={Github}
        query={graphql`
          query AppQuery($login:String!, $first:Int!) {
            organization(login:$login) {
              login
              repositories(first:$first) {
                nodes {
                  name
                }
              }
            }  
          }
        `}
        variables={{
          login: 'stratumn',
          first: 10,
        }}
        render={({error, props}) => {
          if (error) {
            return <Container><Card>Error!</Card></Container>;
          }
          if (!props) {
            return <Container>
              <Card>
                <Card.Content>
                  <Card.Header>Repos</Card.Header>
                </Card.Content>
                <Card.Content>
                  <Placeholder>
                    <Placeholder.Line />
                    <Placeholder.Line />
                    <Placeholder.Line />
                    <Placeholder.Line />
                    <Placeholder.Line />
                  </Placeholder>
                </Card.Content>
              </Card>
            </Container>;
          }
          return <Container>
            <Card>
              <Card.Content>
                <Card.Header>Repos</Card.Header>
              </Card.Content>
              <Card.Content>
                <List>
                    {props.organization.repositories.nodes.map((repo: any) =>
                      <List.Item key={repo.name}>{repo.name}</List.Item>
                    )}
                </List>
              </Card.Content>
            </Card>
          </Container>;
        }}
      />
    );
  }
}

export default App;
