![Ground Control](https://raw.githubusercontent.com/stratumn/groundcontrol/master/hero.png)

# Ground Control (WIP)

Ground Control is an application to help deal with multi-repository development using a user friendly web interface.

You define workspaces in a YAML file.
A workspace contains multiple projects.
A project corresponds to a branch of a repository.

The Ground Control user interface allows you to perform operations across the projects of a workspace, including:

- Cloning all repositories (defaults to `$PWD/workspaces/WORKSPACE/PROJECT`)
- Check the status of repositories against their origins
- Pull all outdated repositories
- Define workspace wide tasks
- Create scripts to launch multi-repository applications

## Installation

You need:

- Go
- Dep
- Node
- Yarn

Clone to `$GOPATH/src/github.com/stratumn/groundcontrol` (no Go module support yet).

Run:

```
make # <---- builds `./groundcontrols`
make install # <---- copies it to `$GOPATH/bin`
```

## Usage

After installing, run:

```
ssh-add # <---- to use your SSH key for accessing private repos
groundcontrol [groundcontrol.yml] # <---- path to YAML file
```

Navigate to `http://localhost:8080`.

## Development

If you didn't run `make`, do:

```
make deps
go generate ./...
```

### Server

```
go run main.go [groundcontrol.yml]
```

Server runs on `http://localhost:8080` and serves GraphiQL instead of the UI during development.

### UI

```
cd ui
yarn dev
```

UI runs on `http://localhost:3000` during development.

### Notes

#### General

- Lines of code don't matter if the code can be generated. Focus on writing repeatable code then write a generator once it makes sense.

#### Server

- All models must be *pure*. Mutations must create a new copy of the model. Avoid pointer accessors for models.
- Reference models by ID. Do not keep pointers. Do not pass models around.
- Models should only store GraphQL fields and IDs of nodes they reference. They must not store a mutex. Use `NodeManager.Lock()`.
- Queries must have no side effects (except appending the log). This may require more mutations triggered by the client.

#### Client

- Queries, mutations, subscriptions, and events must be handled at the top level.
- Avoid using `@relay(mask: false)`.
