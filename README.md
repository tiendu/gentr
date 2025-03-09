**gentr** is a lightweight file-watching utility written in Go. Inspired by the classic [entr](https://github.com/eradman/entr) tool, gentr monitors a list of files (or directories) for changes and automatically executes a specified command when a change is detected. It supports flexible file/directory input, recursive watching, placeholder substitution, structured logging, graceful shutdown, and debouncing.

## Features

### Flexible File Input

- Input Source:

Specify files or directories using the `--input` flag. The input can be a single file, a directory, or a glob pattern (e.g., `logs/*.log`).

- STDIN Override:

If you pipe a list of files via STDIN, gentr will use that list instead of the value provided via `--input`.

### File Watching

- Monitoring:

gentr watches the provided list of files for modifications.

- Recursive Mode:

With the `--recursive` flag enabled, gentr will traverse directories (using Go's `filepath.Walk`) to monitor all files within subdirectories.

### Command Execution with Placeholder

When a change is detected, gentr executes a user-specified command. Use the placeholder `/_` in your command to substitute the changed file's name.

For example, running: `cat /_` and if `testdir/file1.txt` changes, the command will be executed as: `cat testdir/file1.txt`

### Structured Logging

gentr prints both the raw command output and a structured status log (e.g., `exit|0|cat testdir/file1.txt`).

### Graceful Shutdown

gentr listens for SIGINT/SIGTERM signals to exit cleanly.

### Debouncing

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
│   ├── beautify
│   │   └── beautify.go  # ANSI formatting utilities
│   ├── executor.go  # Handles command execution
│   ├── logfilter.go  # Processes the command output, formatting structured logs that indicate command status
│   └── watcher.go  # Watches files (or directories, when recursive mode is enabled) for changes, debounces rapid events, and triggers command execution upon changes
│   ├── resolver.go  # Resolves the input path (file, directory, or glob) for monitoring
├── LICENSE
├── main.go
└── README.md
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

**gentr** reads a list of files from STDIN or the `--input` flag and executes a command when one of the files changes.

For example, to watch all files in a directory recursively and run `cat` on the changed file:

```shell
find testdir -type f | ./gentr --recursive cat /_
```

```shell
./gentr --input testdir --recursive cat /_
```

### Administrative Commands

* Version: `./gentr version`

* Bump Version: `./gentr bump`

### Options

* `--input`

Select a directory/file for monitoring.

* `--debug`

Enable debug mode for more verbose output.

* `--recursive`

Watch directories/files recursively when enabled.
