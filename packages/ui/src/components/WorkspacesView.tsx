import { Link } from "found";
import React, { Component } from "react";
import ReactMarkdown from "react-markdown";
import Moment from "react-moment";
import {
  Button,
  Card,
  Divider,
  Dropdown,
  Feed,
  Header,
  Icon,
  Label,
  Menu,
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
                {project.description || "No description."}
              </Card.Description>
              <Divider horizontal={true}>
                <Header as="h6">Latest Commits</Header>
              </Divider>
              <Feed>
                {events}
              </Feed>
          </Card.Content>
          <Card.Content extra={true}>
            <div className="ui three buttons">
              <Button color="teal" disabled={true}>
                Pull
              </Button>
            </div>
          </Card.Content>
        </Card>
      );
    });

    const notes = workspace.notes || "No notes";

    return (
      <div>
        <Header as="h1">
          <Icon name="cube" />
          <Header.Content>
            {workspace.name}
            <Header.Subheader>
              {workspace.description}
            </Header.Subheader>
          </Header.Content>
        </Header>
        <Label size="large">not cloned</Label>
        <div style={{ margin: "2em 0" }}>
          <ReactMarkdown source={notes} />
        </div>
        <Menu secondary={true}>
          <Menu.Item>
            <Icon name="clone" />
            Clone All
          </Menu.Item>
          <Menu.Item disabled={true}>
            <Icon name="download" />
            Pull Outdated
          </Menu.Item>
          <Menu.Item disabled={true}>
            <Icon name="power" />
            Power On
          </Menu.Item>
          <Menu.Item>
            <Dropdown item={true} text="Tasks" pointing={true} disabled={true}>
              <Dropdown.Menu>
                <Dropdown.Item>Run Tests</Dropdown.Item>
                <Dropdown.Item>Clear Database</Dropdown.Item>
              </Dropdown.Menu>
            </Dropdown>
          </Menu.Item>
        </Menu>
        <Card.Group itemsPerRow={3}>
          {cards}
        </Card.Group>
      </div>
    );
  }
}
