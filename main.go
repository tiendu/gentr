package main

import (
	"io"
	"os"

	"github.com/tiendu/gentr/internal/app"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin *os.File, stdout, stderr io.Writer) int {
	return app.Run(args, stdin, stdout, stderr)
}
