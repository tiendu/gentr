**gentr** is a lightweight file-watching utility written in Go. Inspired by the classic [entr](https://github.com/eradman/entr) tool, gentr monitors a list of files (or directories) for changes and automatically executes a specified command when a change is detected. It supports recursive file watching, placeholder substitution, structured logging, and graceful shutdown.

## Features

### File Watching:

gentr watches a list of files for modifications. With the `--recursive` flag, gentr scans directories recursively and monitors all files.

### Command Execution with Placeholder:

When a change is detected, gentr executes a user-specified command. Use the placeholder `/_` in your command to substitute the changed file's name.

For example, running: `cat /_` and if `testdir/file1.txt` changes, the command will be executed as: `cat testdir/file1.txt`

### Structured Logging:

gentr prints both the raw command output and a structured status log (e.g., `exit|0|cat testdir/file1.txt`).

### Graceful Shutdown:

gentr listens for SIGINT/SIGTERM signals to exit cleanly.

### Debouncing:

It uses a debounce timer (default 500ms) to prevent rapid successive changes from triggering multiple command executions.

## Directory Structure

```
.
├── cmd  # Parse command-line flags, handle administrative subcommands (e.g., version, bump), and display version information.
│   ├── options.go
│   └── version.go
├── entr
├── go.mod
├── internal
│   ├── executor.go  # Handles command execution
│   ├── logfilter.go  # Processes the command output, formatting structured logs that indicate command status
│   └── watcher.go  # Watches files (or directories, when recursive mode is enabled) for changes, debounces rapid events, and triggers command execution upon changes
└── main.go
```

## Installation

1. Clone the Repository

2. Build the Binary

```
go build -o gentr .
```

3. (Optional) Install Globally:

Copy the binary to a directory in your PATH (e.g., `/usr/local/bin`): `sudo cp gentr /usr/local/bin/`

## Usage

### Basic Usage

**gentr** reads a list of files from STDIN and executes a command when one of the files changes.

For example, to watch all files in a directory recursively and run cat on the changed file:

```shell
find testdir -type f | ./gentr --recursive cat /_
```

### Administrative Commands

* Version: `./gentr version`

* Bump Version: `./gentr bump`

### Options

* `--debug`

Enable debug mode for more verbose output.

* `--recursive`

Watch directories recursively. When enabled, gentr traverses directories (using Go’s `filepath.Walk`) and watches all files found.
