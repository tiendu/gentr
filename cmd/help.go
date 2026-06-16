package cmd

import "fmt"

func HelpCommand() int {
	fmt.Println(`Usage: gentr [options] <command>
       gentr <admin-command> [options]

Admin commands:
  version      Print version
  help         Show this message

Watch options:
  --debug, -d        Enable debug mode
  --recursive, -r    Watch directories recursively
  --length, -l       Limit output lines
  --log              Enable logging
  --input, -i        Input path or glob pattern

Examples:
  gentr --input logs/*.log --recursive 'echo changed /_'
  find testdir -type f | gentr --recursive cat /_
  `)
	return 0
}
