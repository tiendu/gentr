# gentr

**gentr** is a lightweight file-watching utility written in Go. It is inspired by the classic [`entr`](https://github.com/eradman/entr) tool.

gentr watches files or directories and runs a command when something changes. It supports file/directory input, glob input, recursive watching, placeholder substitution, structured output, optional logging, graceful shutdown, and debouncing.

## Features

### Flexible input

Use `--input` to watch a file, directory, or glob pattern:

```shell
gentr --input 'logs/*.log' cat /_
```

You can also pipe files through STDIN. When STDIN is provided, it takes priority over `--input`:

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

gentr prints raw command output and a status line such as:

```text
exit|0|cat testdir/file1.txt
```

### Optional logging

Use `--log` to write change records and command status to a timestamped log file.

### Graceful shutdown

gentr listens for `SIGINT` and `SIGTERM` and shuts down cleanly.

## Design

The code uses small Go interfaces and composition instead of inheritance-heavy structure.

Core boundaries:

- `Resolver` resolves files from an input path or glob.
- `StdinReader` reads file paths from STDIN.
- `CommandRunner` executes the user command.
- `OutputReporter` prints command output/status.
- `ChangeLogger` writes optional session logs.
- `Spinner` controls terminal activity display.
- `Watcher` coordinates polling, debouncing, command execution, diff rendering, and logging.

This keeps the CLI easy to test and avoids tying the internal package to the command parser.

## Directory structure

```text
.
├── cmd
│   ├── flags.go      # CLI option parsing
│   ├── help.go       # Help text
│   ├── install.go    # install/uninstall/reinstall commands
│   ├── router.go     # Small command router interface
│   ├── spinner.go    # Spinner interface implementations
│   └── version.go    # Version command
├── internal
│   ├── config.go     # WatchOptions
│   ├── diff.go       # Line diff helpers
│   ├── executor.go   # CommandRunner interface + shell runner
│   ├── logger.go     # OutputReporter and ChangeLogger interfaces
│   ├── resolver.go   # Resolver and StdinReader interfaces
│   ├── watcher.go    # Watcher orchestration
│   └── utils
│       └── utils.go  # Terminal formatting helpers
├── go.mod
├── LICENSE
├── main.go
└── README.md
```

## Installation

Build locally:

```shell
go build -o gentr .
```

Optionally install globally:

```shell
sudo cp gentr /usr/local/bin/
```

Or use the built-in installer:

```shell
gentr install
```

The default install path is:

```text
~/.local/bin
```

You can override it:

```shell
INSTALL_PATH=/usr/local/bin gentr install
```

## Usage

Watch a directory recursively and run `cat` on the changed file:

```shell
gentr --input testdir --recursive cat /_
```

Use STDIN:

```shell
find testdir -type f | gentr cat /_
```

Limit command output lines:

```shell
gentr --input testdir --recursive --length 20 'pytest /_'
```

Enable logging:

```shell
gentr --input testdir --recursive --log 'echo changed /_'
```

## Admin commands

```shell
gentr version
gentr install
gentr uninstall
gentr reinstall
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
