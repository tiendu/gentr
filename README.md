# gentr

**gentr** is a lightweight file-watching utility written in Go. It is inspired by the classic [`entr`](https://github.com/eradman/entr) tool.

gentr watches files or directories and runs a command when something changes. It supports file and directory input, glob patterns, recursive watching, placeholder substitution, structured output, optional logging, graceful shutdown, and debouncing.

## Features

### Flexible input

Use `--input` to watch a file, directory, or glob pattern:

```shell
gentr --input 'logs/*.log' cat /_
```

You can also pipe file paths through standard input. When standard input is provided, it takes priority over `--input`:

```shell
find testdir -type f | gentr cat /_
```

### Recursive watching

Use `--recursive` to watch files inside subdirectories:

```shell
gentr --input testdir --recursive cat /_
```

### Placeholder substitution

Use `/_` inside the command to represent the changed file:

```shell
gentr --input testdir --recursive 'echo changed /_'
```

If `testdir/file1.txt` changes, gentr runs:

```shell
echo changed testdir/file1.txt
```

### Structured output

gentr prints raw command output followed by a status line:

```text
exit|0|cat testdir/file1.txt
```

### Optional logging

Use `--log` to write change records and command status to a timestamped log file.

### Graceful shutdown

gentr listens for `SIGINT` and `SIGTERM` and shuts down cleanly.

## Design

The code uses small Go interfaces and composition rather than inheritance-heavy structure.

Core boundaries:

- `Resolver` resolves files from an input path or glob.
- `StdinReader` reads file paths from standard input.
- `CommandRunner` executes the user command.
- `OutputReporter` prints command output and status.
- `ChangeLogger` writes optional session logs.
- `Spinner` controls terminal activity display.
- `Watcher` coordinates polling, debouncing, command execution, diff rendering, and logging.

This keeps the CLI easy to test and avoids coupling the internal package to command-line parsing.

Build, installation, removal, testing, and cleanup are handled by the `Makefile`. The gentr binary only contains application behavior.

## Directory structure

```text
.
├── cmd
│   ├── flags.go          # CLI option parsing
│   ├── flags_test.go
│   ├── help.go           # Help text
│   ├── router.go         # Small command router
│   ├── router_test.go
│   ├── spinner.go        # Spinner interface implementations
│   ├── spinner_test.go
│   └── version.go        # Version command
├── internal
│   ├── config.go         # WatchOptions
│   ├── diff.go           # Line diff helpers
│   ├── diff_test.go
│   ├── executor.go       # CommandRunner interface and shell runner
│   ├── executor_test.go
│   ├── logger.go         # OutputReporter and ChangeLogger interfaces
│   ├── logger_test.go
│   ├── resolver.go       # Resolver and StdinReader interfaces
│   ├── resolver_test.go
│   ├── watcher.go        # Watcher orchestration
│   ├── watcher_test.go
│   └── utils
│       ├── utils.go      # Terminal formatting helpers
│       └── utils_test.go
├── .gitignore
├── go.mod
├── LICENSE
├── main.go
├── main_test.go
├── Makefile
└── README.md
```

## Requirements

- Go 1.23 or newer
- `make` for the Makefile workflow

## Installation

### Install with Make

Clone the repository and install the binary to `~/.local/bin`:

```shell
git clone https://github.com/tiendu/gentr.git
cd gentr
make install
```

The default installation path is:

```text
~/.local/bin/gentr
```

Make sure `~/.local/bin` is in your `PATH`.

For zsh:

```shell
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

For bash:

```shell
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Install to another location

Set `PREFIX`:

```shell
make install PREFIX=/usr/local
```

This installs the binary to:

```text
/usr/local/bin/gentr
```

You can also set the full binary directory directly:

```shell
make install BIN_DIR="$HOME/bin"
```

### Install with Go

You can install gentr directly with Go:

```shell
go install github.com/tiendu/gentr@latest
```

The binary is installed into `GOBIN`, or into `$(go env GOPATH)/bin` when `GOBIN` is not set.

### Build without installing

```shell
make build
```

The binary is written to:

```text
build/gentr
```

Run it directly:

```shell
./build/gentr --input testdir --recursive cat /_
```

## Development

Run all validation checks:

```shell
make check
```

Run tests:

```shell
make test
```

Run tests with the race detector:

```shell
make test-race
```

Run static analysis:

```shell
make vet
```

Format the source:

```shell
make fmt
```

Build the binary:

```shell
make build
```

Remove build and coverage output:

```shell
make clean
```

Uninstall the binary:

```shell
make uninstall
```

To uninstall from a custom prefix, use the same value used during installation:

```shell
make uninstall PREFIX=/usr/local
```

## Usage

Watch a directory recursively and run `cat` on the changed file:

```shell
gentr --input testdir --recursive cat /_
```

Use standard input:

```shell
find testdir -type f | gentr cat /_
```

Watch Go files and rerun tests:

```shell
gentr --input . --recursive go test ./...
```

Watch a glob pattern:

```shell
gentr --input '*.go' go test ./...
```

Limit command output lines:

```shell
gentr --input testdir --recursive --length 20 'pytest /_'
```

Enable logging:

```shell
gentr --input testdir --recursive --log 'echo changed /_'
```

## Commands

```shell
gentr version
gentr help
```

Installation and removal are intentionally handled by the Makefile rather than the running binary:

```shell
make install
make uninstall
```

## Options

```text
--debug, -d        Enable debug mode
--recursive, -r    Watch directories recursively
--length, -l       Limit output lines
--log              Enable logging
--input, -i        Input path or glob pattern
```

## License

See [LICENSE](LICENSE).
