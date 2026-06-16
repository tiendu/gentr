package cmd

import "fmt"

var VersionStr = "0.3.0"
var RevisionStr = "unknown"

func VersionCommand() int {
	fmt.Printf("gentr version %s, build revision %s\n", VersionStr, RevisionStr)
	return 0
}
