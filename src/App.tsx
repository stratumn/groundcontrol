import React, { Component } from 'react';
import {QueryRenderer} from 'react-relay';
/// <reference path="declarations.d.ts"/>
import graphql from 'babel-plugin-relay/macro';
import {Github} from './relay';
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
            return <div>Error!</div>;
          }
          if (!props) {
            return <div>Loading...</div>;
          }
          console.log(props);
          return <div>
            <h1>{props.organization.login}</h1>
            <ul>
               {props.organization.repositories.nodes.map((repo: any) =>
                 <li key={repo.name}>{repo.name}</li>
               )}
            </ul>
          </div>
        }}
      />
    );
  }
}

export default App;
