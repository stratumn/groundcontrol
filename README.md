# Ground Control (WIP)

_Ground Control_ is an application to help deal with multi-repository development using a user friendly web interface.

![Ground Control](https://raw.githubusercontent.com/stratumn/groundcontrol/master/screenshot.png)

Workspaces are defined using YAML files which can easily be shared.
Each workspace contains multiple projects.
A project corresponds to a Git reference of a repository (such as a branch or tag).

The user interface allows you to perform operations on multiple projects at once, including:

- Automatically sync and share workspaces using _sources_
- Clone all the repositories in a workspace (defaults to `$HOME/groundcontrol/workspaces/$WORKSPACE/$PROJECT`)
- See if you are up-to-date or ahead of the remote repositories
- Run tasks on multiple repositories
- Launch services and their dependencies with ease

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

### Windows (scoop)

Windows support is **experimental**.

You need Git and possibly Putty with Pageant configured to use your Git ssh keys.

```
scoop bucket add groundcontrol https://github.com/stratumn/groundcontrol-scoop-bucket.git
scoop install groundcontrol
```

Once installed, you can update to latest version by running:

```
scoop update groundcontrol
```

### Prebuilt binaries

Head over to the [latest release](https://github.com/stratumn/groundcontrol/releases/latest) and download the binary for your platform.

## Usage

After installing, run:

```
ssh-add # <---- may be needed in order to use your SSH key for accessing repos
groundcontrol
```

## Development

Use this source:

```
git@github.com:stratumn/groundcontrol-source.git
```
