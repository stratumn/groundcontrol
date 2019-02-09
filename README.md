![Ground Control](https://raw.githubusercontent.com/stratumn/groundcontrol/master/hero.png)

# Ground Control (WIP)

Ground Control is an application to help deal with multi-repository development using a user friendly web interface.

You define workspaces in a YAML file.
A workspace contains multiple projects.
A project corresponds to a branch of a repository.

The Ground Control user interface allows you to perform operations across the projects of a workspace, including:

- Clone all repositories (defaults to `$HOME/groundcontrol/workspaces/WORKSPACE/PROJECT`)
- Check the status of repositories against their origins
- Pull all outdated repositories
- Define workspace wide tasks
- Create scripts to launch multi-repository applications

![Ground Control](https://raw.githubusercontent.com/stratumn/groundcontrol/master/screenshot.png)

## Installation

### macOS (homebrew)

Simply run:

```
brew install stratumn/groundcontrol/groundcontrol
```

Once installed, you can update to latest version by running:

```
brew upgrade groundcontrol
```

### Prebuilt binaries

Head over to the [latest release](https://groundcontrol/releases/latest) and download the binary for your platform.

### From source

You need:

- Go >= v1.11
- Node
- Yarn

Clone outside of `$GOPATH` since it's a Go module.

Run:

```
make # <---- builds `./groundcontrol`
make install # <---- copies it to `$GOPATH/bin`
```

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
