import graphql from "babel-plugin-relay/macro";
import debounce from "debounce";
import { Link } from "found";
import React, { Component } from "react";
import { commitMutation, createFragmentContainer, RelayProp } from "react-relay";
import {
  Button,
  Card,
  Divider,
  Header,
  Icon,
  Input,
  InputOnChangeData,
  Label,
  List,
  Placeholder,
 } from "semantic-ui-react";

import { RouterWorkspacesListQueryResponse } from "../__generated__/RouterWorkspacesListQuery.graphql";

import RepoShortName from "./RepoShortName";

interface IProps extends RouterWorkspacesListQueryResponse {
  relay: RelayProp;
}

interface IState {
  searchQuery: string;
}

const cloneWorkspaceMutation = graphql`
  mutation WorkspacesListCloneWorkspaceMutation($id: String!) {
    cloneWorkspace(id: $id) {
      id
    }
  }
`;

export class WorkspacesList extends Component<IProps, IState> {

  private handleSearchChange = debounce((e: React.ChangeEvent<HTMLInputElement>, { value }: InputOnChangeData) => {
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
      let workspaces = this.props.viewer.workspaces!;

      if (this.state.searchQuery.length > 0) {
        workspaces = workspaces.filter((elem) => (
          elem!.name.toLowerCase().indexOf(this.state.searchQuery.toLowerCase()) >= 0
        ));
      }

      cards = workspaces.map((workspace) => {
        const items = workspace!.projects!.map((project) => (
          <List.Item key={project.id}>
            <List.Content floated="right">
              <Label
                style={{ position: "relative", top: "-.3em" }}
                size="small"
              >
                {project.branch}
              </Label>
            </List.Content>
            <List.Content>
              <RepoShortName repo={project.repo} />
            </List.Content>
          </List.Item>
        ));

        return (
          <Card key={workspace!.id}>
            <Card.Content>
              <Link
                to={`/workspaces/${workspace!.slug}`}
                Component={Card.Header}
              >
                {workspace!.name}
              </Link>
              <Card.Meta>
                {workspace!.description}
              </Card.Meta>
              <Divider horizontal={true}>
                <Header as="h6">Repositories</Header>
              </Divider>
              <Card.Description style={{ marginTop: "1em" }}>
                <List>
                  {items}
                </List>
              </Card.Description>
            </Card.Content>
            <Card.Content extra={true}>
              <div className="ui three buttons">
                <Link
                  to={`/workspaces/${workspace!.slug}`}
                  className="ui grey button"
                >
                  Details
                </Link>
                <Button
                  color="teal"
                  onClick={this.handleCloneWorkspace.bind(this, workspace!.id)}
                >
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
              A workspace is a collection of related projects. Each project is linked to a Github repository and branch.
            </Header.Subheader>
          </Header.Content>
        </Header>
        <Input
          size="huge"
          icon="search"
          placeholder="Search..."
          style={{marginBottom: "2em"}}
          onChange={this.handleSearchChange}
         />
        <Card.Group itemsPerRow={3}>
          {cards}
        </Card.Group>
      </div>
    );
  }

  private handleCloneWorkspace(id: string) {
    commitMutation(this.props.relay.environment, {
      mutation: cloneWorkspaceMutation,
      variables: {
        id,
      },
    });
  }
}

export default createFragmentContainer(WorkspacesList, {});
