import { Link } from "found";
import debounce from "debounce";
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

interface Props extends RouterWorkspacesListQueryResponse {
}

interface State {
  searchQuery: string
}

export default class WorkspacesList extends Component<Props, State> {

  constructor(props: Props) {
    super(props);
    this.state = { searchQuery: "" };
  }

  render() {
    if (!this.props) {
      return <div>
        <Card.Group>
          {[...Array(10)].map((_, i) =>
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
          )}
        </Card.Group>
      </div>;
    }

    let workspaces = this.props.workspaces!;

    if (this.state.searchQuery.length > 0) {
      workspaces = workspaces.filter((elem) =>
        elem.name.toLowerCase().indexOf(this.state.searchQuery.toLowerCase()) >= 0
      );
    }

    return <div>
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
        {workspaces!.map((workspace: any) =>
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
                  {workspace.projects.map((project: any) =>
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
                  )}
                </List>
              </Card.Description>
            </Card.Content>
            <Card.Content extra>
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
        )}
      </Card.Group>
    </div>;
  }

  handleSearchChange = debounce((e: React.MouseEvent<HTMLElement>, { value }: SearchProps) => {
    this.setState({ searchQuery: value as string });
  }, 100);
};
