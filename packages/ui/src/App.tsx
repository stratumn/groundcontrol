import React, { Component } from 'react';
import { QueryRenderer } from 'react-relay';
import graphql from 'babel-plugin-relay/macro';
import GroundControl from './groundcontrol.env.relay';
import { Container, Card, Placeholder, List } from 'semantic-ui-react';
import './App.css';

class App extends Component {
  render() {
    return (
      <QueryRenderer
        environment={GroundControl}
        query={graphql`
          query AppQuery {
            workspaces {
              name
              slug
            }  
          }
        `}
        variables={{}}
        render={({error, props}) => {
          if (error) {
            return <Container><Card>Error!</Card></Container>;
          }
          if (!props) {
            return <Container>
              <Card>
                <Card.Content>
                  <Card.Header>Workspaces</Card.Header>
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
                <Card.Header>Workspaces</Card.Header>
              </Card.Content>
              <Card.Content>
                <List>
                    {props.workspaces.map((workspace: any) =>
                      <List.Item key={workspace.slug}>{workspace.name}</List.Item>
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
