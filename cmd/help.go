package cmd

import "fmt"

func HelpCommand() {
	fmt.Println(`Usage: gentr <command> [options]

Commands:
  install      Install this tool as a system command
  uninstall    Remove the installed tool
  reinstall    Reinstall the tool
  version      Print version
  bump         Bump the version number
  help         Show this message

Options:
  --debug, -d        Enable debug mode
  --recursive, -r    Watch directories recursively
  --length, -l       Limit output lines
  --log              Enable logging
  --input, -i        Input path or glob pattern

Examples:
  gentr --input logs/*.log --recursive 'echo changed /_'
  INSTALL_PATH=/usr/local/bin gentr install
  `)
}
