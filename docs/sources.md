# Sources

Sources contain YAML files that define workspaces. When loading a source,
Ground Control looks for workspaces in all the YAML files within the directory.
It's up to you to decide whether to put them all in one file, or to split them.

## Creating a Source

The easiest way to create a new source is to create a directory and add it as a
Directory Source in Ground Control. Just because it is a directory doesn't mean
you can't initialize a Git repository for it. Setting up Git will allow other
people to use it as a Git Source.

Note that currently sources are synced about once every minute by default, so it
takes a little bit of time before the workspaces are refreshed. In the future
Ground Control will be able to watch changes to files, and reload workspaces
quicker.

## YAML

Each YAML file contains one or more workspaces:

```yaml
workspaces:
  - name: My Workspace
    # slug should be a globally unique, URL friendly identifier
    slug: my-workspace
    description: A workspace containing two projects.
    notes: |
      ## Notes
      Here you can use markdown to add notes to your workspace.
    projects:
        # slug should be a unique identifier within the scope of the workspace
      - slug: backend
        repository: git@github.com:user/backend.git
        # reference is any Git reference, such as a branch or a tag
        reference: refs/heads/master
        description: The backend for the application.
      - slug: frontend
        repository: git@github.com:user/frontent.git
        reference: refs/heads/master
        description: The frontend for the application.
    services:
      - name: Backend
        # variables can be defined at the service or task level
        variables:
            # an environment variable with the name of the variable will be set
          - name: BACKEND_PORT
            # they can have a default value
            default: 3000
        # if project contains the slug of the project, the command will run in that project
        project: backend
        # the command will be executed in a shell
        command: PORT=$BACKEND_PORT npm start
        # before/after lists names of tasks to run before/after running the command
        before:
          - Install Backend Dependencies
      - name: Frontend
        variables:
          - name: FRONTEND_PORT
            default: 4000
          - name: BACKEND_PORT
        project: frontend
        command: PORT=$FRONTEND_PORT BACKEND_PORT=$BACKEND_PORT npm start
        before:
          - Install Frontend Dependencies
        # needs list names of services that must be started before this one
        needs:
          - Backend
    tasks:
      - name: Install Backend Dependencies
        # a task can contain multiple steps
        steps:
            # each step is executed on one or more projects
          - projects:
              - backend
            # unlike services, tasks can run multiple shell commands
            commands:
              - npm install
      - name: Install Frontend Dependencies
        steps:
          - projects:
              - frontend
            commands:
              - npm install
```

Commands are executed using an embedded sh-like shell.