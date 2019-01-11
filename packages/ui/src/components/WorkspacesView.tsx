import { Link } from "found";
import React, { Component } from "react";
import Moment from "react-moment";
import {
  Button,
  Card,
  Feed,
  Header,
  Icon,
  Label,
  List,
 } from "semantic-ui-react";

import { RouterWorkspacesViewQueryResponse } from "../__generated__/RouterWorkspacesViewQuery.graphql";

type IProps = RouterWorkspacesViewQueryResponse;

export default class WorkspacesView extends Component<IProps> {

  public render() {
    if (!this.props) {
      return <div>Loading...</div>;
    } else if (!this.props.workspace) {
      return <div>Not found.</div>;
    }
    const workspace = this.props.workspace!;

    const cards = workspace.projects!.map((project) => {
      const events = project.commits.edges!.map((commit) => (
        <Feed.Event key={commit!.node.id}>
          <Feed.Content>
            <Feed.Summary>
              {commit.node.headline}
            </Feed.Summary>
            <Feed.Meta>
              Pushed by <strong>{commit!.node.author}</strong>
              <Moment
                fromNow={true}
                style={{marginLeft: 0}}
              >
                {commit!.node.date}
              </Moment>
            </Feed.Meta>
          </Feed.Content>
        </Feed.Event>
      ));

      return (
        <Card key={project.repo}>
          <Card.Content>
            <Card.Header>{project.repo.replace("github.com/", "")}</Card.Header>
              <Label style={{ marginTop: ".8em" }}>{project.branch}</Label>
              <Card.Description style={{ marginTop: "1em" }}>
              <p>{project.description || "No description."}</p>
              <Feed>
                {events}
              </Feed>
            </Card.Description>
          </Card.Content>
          <Card.Content extra={true}>
            <div className="ui three buttons">
              <Button color="teal">
                Pull
              </Button>
            </div>
          </Card.Content>
        </Card>
      );
    });

    return (
      <div>
        <Header as="h1" style={{ marginBottom: "1.2em" }} >
          <Icon name="cube" />
          <Header.Content>
            {workspace.name}
            <Header.Subheader>
              {workspace.description}
            </Header.Subheader>
          </Header.Content>
        </Header>
        <p style={{ marginBottom: "2em" }}>
          {workspace.notes || "No notes."}
        </p>
        <Card.Group>
          {cards}
        </Card.Group>
      </div>
    );
  }
}
