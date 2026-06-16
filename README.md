# gentr

**gentr** is a lightweight file-watching utility written in Go. It is inspired by the classic [`entr`](https://github.com/eradman/entr) tool.

gentr watches files or directories and runs a command when something changes. It supports file and directory input, glob patterns, recursive watching, placeholder substitution, structured output, optional logging, graceful shutdown, and debouncing.

## Features

### Flexible input

Use `--input` to watch a file, directory, or glob pattern:

```shell
gentr --input 'logs/*.log' cat /_
```

You can also pipe file paths through standard input. Piped input takes priority over `--input`:

```shell
find testdir -type f | gentr cat /_
```

### Recursive watching

```shell
gentr --input testdir --recursive cat /_
```

### Placeholder substitution

Use `/_` to represent the changed file:

```shell
gentr --input testdir --recursive 'echo changed /_'
```

### Structured output

```text
exit|0|cat testdir/file1.txt
```

### Optional logging

Use `--log` to write change records and command status to a timestamped log file.

### Graceful shutdown

gentr listens for `SIGINT` and `SIGTERM` and shuts down cleanly.

## Design

The project is split into small, cohesive internal packages. Each package keeps its tests beside the implementation it validates, following normal Go conventions.

```text
.
├── internal
│   ├── app
│   │   ├── app.go
│   │   └── app_test.go
│   ├── buildinfo
│   │   └── buildinfo.go
│   ├── cli
│   │   ├── cli.go
│   │   └── cli_test.go
│   ├── config
│   │   ├── options.go
│   │   └── options_test.go
│   ├── diff
│   │   ├── diff.go
│   │   └── diff_test.go
│   ├── input
│   │   ├── resolver.go
│   │   └── resolver_test.go
│   ├── output
│   │   ├── output.go
│   │   └── output_test.go
│   ├── runner
│   │   ├── runner.go
│   │   └── runner_test.go
│   ├── spinner
│   │   ├── spinner.go
│   │   └── spinner_test.go
│   ├── terminal
│   │   ├── terminal.go
│   │   └── terminal_test.go
│   └── watch
│       ├── watcher.go
│       └── watcher_test.go
├── .gitignore
├── go.mod
├── LICENSE
├── main.go
├── main_test.go
├── Makefile
└── README.md
```

Core boundaries are expressed as small interfaces where they are consumed:

- `Resolver` discovers files from a path or glob.
- `StdinReader` reads paths from standard input.
- `CommandRunner` executes commands.
- `OutputReporter` renders command output.
- `ChangeLogger` stores optional session records.
- `Spinner` controls terminal activity display.

`main.go` only delegates to the application package. Build, test, install, uninstall, and cleanup are handled by the Makefile.

## Requirements

- Go 1.23 or newer
- `make` for the Makefile workflow

## Installation

```shell
git clone https://github.com/tiendu/gentr.git
cd gentr
make install
```

The default destination is:

```text
~/.local/bin/gentr
```

Custom destination:

```shell
make install PREFIX=/usr/local
make install BIN_DIR="$HOME/bin"
```

Install through Go:

```shell
go install github.com/tiendu/gentr@latest
```

## Development

```shell
make check
make test
make test-race
make build
make clean
```

Uninstall:

```shell
make uninstall
```

## Usage

```shell
gentr --input testdir --recursive cat /_
find testdir -type f | gentr cat /_
gentr --input . --recursive go test ./...
gentr --input '*.go' go test ./...
gentr --input testdir --recursive --length 20 'pytest /_'
gentr --input testdir --recursive --log 'echo changed /_'
```

## Commands

```shell
gentr version
gentr help
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
