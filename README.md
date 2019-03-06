![Ground Control](https://raw.githubusercontent.com/stratumn/groundcontrol/master/hero.png)

# Ground Control (WIP)

_Ground Control_ is an application to help deal with multi-repository development using a user friendly web interface.

Workspaces are defined using YAML files which can easily be shared.
Each workspace contains multiple projects.
A project corresponds to a Git reference of a repository (such as a branch or tag).

The user interface allows you to perform operations on multiple projects at once, including:

- Automatically sync and share workspaces using _sources_
- Clone all the repositories in a workspace (defaults to `$HOME/groundcontrol/workspaces/$WORKSPACE/$PROJECT`)
- See if you are up-to-date or ahead of the remote repositories
- Run tasks on multiple repositories
- Launch services and their dependencies with ease

![Ground Control](https://raw.githubusercontent.com/stratumn/groundcontrol/master/screenshot.png)

## Installation

### macOS (homebrew)

Simply run:

```
brew install stratumn/groundcontrol/groundcontrol
```

Once installed, you can update to latest version by running:

```
brew upgrade stratumn/groundcontrol/groundcontrol
```

### Prebuilt binaries

Head over to the [latest release](https://github.com/stratumn/groundcontrol/releases/latest) and download the binary for your platform.

## Usage

After installing, run:

```
ssh-add # <---- to use your SSH key for accessing private repos
groundcontrol
```

## Development

Use this source:

```
git@github.com:stratumn/groundcontrol-source.git
```
