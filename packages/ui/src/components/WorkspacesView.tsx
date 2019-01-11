import { Link } from "found";
import debounce from "debounce";
import React, { Component } from "react";
import {
  Button,
  Card,
  Feed,
  Header,
  Icon,
  Label,
 } from "semantic-ui-react";

import { RouterWorkspacesListQueryResponse } from "../__generated__/RouterWorkspacesListQuery.graphql";

interface Props extends RouterWorkspacesListQueryResponse {
}

export default class WorkspacesView extends Component<Props> {

  render() {
    return <div>
      <Header as="h1" style={{ marginBottom: "1.2em" }} >
        <Icon name="cube" />
        <Header.Content>
          Workspace name
          <Header.Subheader>
            Workspace description.
          </Header.Subheader>
        </Header.Content>
      </Header>
      <p style={{ marginBottom: "2em" }}>Workspace notes for example installation notes.</p>
      <Card.Group>
        <Card>
          <Card.Content>
            <Card.Header>stratumn/repo</Card.Header>
              <Label style={{ marginTop: ".8em" }}>branch</Label>
              <Card.Description style={{ marginTop: "1em" }}>
              <p>Project description extracted from github.</p>
              <Feed>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
              </Feed>
            </Card.Description>
          </Card.Content>
          <Card.Content extra>
            <div className="ui three buttons">
              <Button color="teal">
                Pull
              </Button>
            </div>
          </Card.Content>
        </Card>
        <Card>
          <Card.Content>
            <Card.Header>stratumn/repo</Card.Header>
              <Label style={{ marginTop: ".8em" }}>branch</Label>
              <Card.Description style={{ marginTop: "1em" }}>
              <p>Project description extracted from github.</p>
              <Feed>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
              </Feed>
            </Card.Description>
          </Card.Content>
          <Card.Content extra>
            <div className="ui three buttons">
              <Button color="teal">
                Pull
              </Button>
            </div>
          </Card.Content>
        </Card>
        <Card>
          <Card.Content>
            <Card.Header>stratumn/repo</Card.Header>
              <Label style={{ marginTop: ".8em" }}>branch</Label>
              <Card.Description style={{ marginTop: "1em" }}>
              <p>Project description extracted from github.</p>
              <Feed>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
              </Feed>
            </Card.Description>
          </Card.Content>
          <Card.Content extra>
            <div className="ui three buttons">
              <Button color="teal">
                Pull
              </Button>
            </div>
          </Card.Content>
        </Card>
        <Card>
          <Card.Content>
            <Card.Header>stratumn/repo</Card.Header>
              <Label style={{ marginTop: ".8em" }}>branch</Label>
              <Card.Description style={{ marginTop: "1em" }}>
              <p>Project description extracted from github.</p>
              <Feed>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
                <Feed.Event>
                  <Feed.Content>
                    <Feed.Summary>
                      Git commit message headline
                    </Feed.Summary>
                    <Feed.Meta>
                      Pushed by Author 1h ago
                    </Feed.Meta>
                  </Feed.Content>  
                </Feed.Event>
              </Feed>
            </Card.Description>
          </Card.Content>
          <Card.Content extra>
            <div className="ui three buttons">
              <Button color="teal">
                Pull
              </Button>
            </div>
          </Card.Content>
        </Card>
      </Card.Group>
    </div>;
  }
};
