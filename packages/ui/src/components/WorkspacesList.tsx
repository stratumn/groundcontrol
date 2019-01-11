import debounce from "debounce";
import { Link } from "found";
import React, { Component } from "react";
import {
  Button,
  Card,
  Header,
  Icon,
  Label,
  List,
  Placeholder,
  Search,
  SearchProps,
 } from "semantic-ui-react";

import { RouterWorkspacesListQueryResponse } from "../__generated__/RouterWorkspacesListQuery.graphql";

type IProps = RouterWorkspacesListQueryResponse;

interface IState {
  searchQuery: string;
}

export default class WorkspacesList extends Component<IProps, IState> {

  private handleSearchChange = debounce((e: React.MouseEvent<HTMLElement>, { value }: SearchProps) => {
    this.setState({ searchQuery: value as string });
  }, 100);

  constructor(props: IProps) {
    super(props);
    this.state = { searchQuery: "" };
  }

  public render() {
    let cards: JSX.Element[];

    if (!this.props) {
      cards = [...Array(10)].map((_, i) => (
        <Card key={i}>
          <Card.Content>
            <Placeholder>
              <Placeholder.Header>
                <Placeholder.Line />
                <Placeholder.Line />
              </Placeholder.Header>
              <Placeholder.Paragraph>
                <Placeholder.Line length="medium" />
                <Placeholder.Line length="short" />
              </Placeholder.Paragraph>
            </Placeholder>
          </Card.Content>
        </Card>
      ));
    } else {
      let workspaces = this.props.workspaces!;

      if (this.state.searchQuery.length > 0) {
        workspaces = workspaces.filter((elem) => (
          elem.name.toLowerCase().indexOf(this.state.searchQuery.toLowerCase()) >= 0
        ));
      }

      cards = workspaces.map((workspace) => {
        const items = workspace.projects!.map((project) => (
          <List.Item key={project.repo}>
            <List.Content floated="right">
              <Label
                style={{ position: "relative", top: "-.3em" }}
                size="small"
              >
                {project.branch}
              </Label>
            </List.Content>
            <List.Content>
              {project.repo.replace("github.com/", "")}
            </List.Content>
          </List.Item>
        ));

        return (
          <Card key={workspace.slug}>
            <Card.Content>
              <Link
                to={`/workspaces/${workspace.slug}`}
                Component={Card.Header}
              >
                {workspace.name}
              </Link>
              <Card.Meta>
                {workspace.description}
              </Card.Meta>
              <Card.Description style={{ marginTop: "1em" }}>
                <List>
                  {items}
                </List>
              </Card.Description>
            </Card.Content>
            <Card.Content extra={true}>
              <div className="ui three buttons">
                <Link
                  to={`/workspaces/${workspace.slug}`}
                  className="ui grey button"
                >
                  Details
                </Link>
                <Button color="teal">
                  Clone
                </Button>
              </div>
            </Card.Content>
          </Card>
        );
      });
    }

    return (
      <div>
        <Header as="h1" style={{ marginBottom: "1.2em" }} >
          <Icon name="cubes" />
          <Header.Content>
            Workspaces
            <Header.Subheader>
              A workspace is a collection of related projects. Each project is linked to a Github repo and branch.
            </Header.Subheader>
          </Header.Content>
        </Header>
        <Search
          style={{ marginBottom: "2em" }}
          open={false}
          size="large"
          onSearchChange={this.handleSearchChange}
        />
        <Card.Group>
          {cards}
        </Card.Group>
      </div>
    );
  }
}
